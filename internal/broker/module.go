package broker

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"therealbroker/pkg/broker"
	db "therealbroker/pkg/database"
	"therealbroker/pkg/database/dbtype"
	"therealbroker/pkg/prometheus"

	//"therealbroker/pkg/prometheus"
	"time"

	"github.com/gocql/gocql"
)

type Module struct {
	subscribers map[string][]chan broker.Message
	messages    []broker.Message
	messageMutex sync.Mutex
	sync.RWMutex
	isClosed bool
	database db.Database
	checkMsg chan int
	lastMessage int
}

func NewModule() broker.Broker {
	dbProvide := &db.DBProvider{
		DatabaseType: &dbtype.Pgsql{Name: "broker", Type: "sql"},
		// DatabaseType: &dbtype.CassandraDatabase{Name: "broker", Type: "nosql"},
		// DatabaseType: &dbtype.Scylla{Name: "broker", Type: "nosql"},
	}
	db, err2 := dbProvide.Provide()
	if err2 != nil {log.Fatalf("error in shooting database")}
	err := db.Connect()
	if err != nil {log.Fatalf("error in connecting database")}
	return &Module{
		subscribers: make(map[string][]chan broker.Message),
		messages:    make([]broker.Message, 1),
		database: db,
		checkMsg: make(chan int, 1),
		lastMessage: 0,
	}
}

func (m *Module) Close() error {
	m.Lock()
	defer m.Unlock()
	for _, subscribers := range m.subscribers {
		for _, subscriber := range subscribers {
			close(subscriber)
			prometheus.ActiveSubscribers.Dec()
		}
	}
	m.isClosed = true
	return nil
}

func (m *Module) Publish(ctx context.Context, subject string, msg broker.Message) (int, error) {
	if  m.isClosed {return 0, broker.ErrUnavailable}
	m.Lock()
	if checkMsgId := existsInSlice(msg, m.messages); checkMsgId == 0 {
		if len(m.messages) == 0 {
			msg.Id = 1 + m.lastMessage
		} else {
			msg.Id = m.messages[len(m.messages) - 1].Id  + 1
		}
		m.messages = append(m.messages, msg)
	} else {
		msg.Id = checkMsgId + 1
	}
	m.Unlock()

	if len(m.messages) > 3000 {
		m.messageMutex.Lock()
		msg.PublishDate = time.Now()
		timestamp := msg.PublishDate.Format("2006-01-02 15:04:05.999999-07:00")
		m.lastMessage = m.messages[len(m.messages)-1].Id
		if m.database.GetType() == "sql" {
			query := "INSERT INTO messages (body, expiration, publish_date) VALUES ($1, $2, $3)"
			tx, err := m.database.GetDB().(*sql.DB).Begin()
			if err != nil {return 0, err}
			for _, msg := range m.messages {
				params := []interface{}{msg.Body, msg.Expiration.String(), timestamp}
				_, errExec := tx.Exec(query, params...)
				if errExec != nil {
					tx.Rollback()
				}
			}
			if errCommit := tx.Commit(); err != nil {
				fmt.Println("Failed to commit with transaction:", errCommit)
				tx.Rollback()
			}
		} else {
			batch := m.database.GetDB().(*gocql.Session).NewBatch(gocql.LoggedBatch)
			for _, msg := range m.messages {
				nanoTime := time.Now().UnixNano()
				t := gocql.UUIDFromTime(time.Unix(0, nanoTime))
				params := []interface{}{t, msg.Body, msg.Expiration.String(), timestamp}
				query := "INSERT INTO " + m.database.GetName() + "." + "messages" + "(id, body, expiration, publish_date)" + "VALUES (?,?,?,?)" 
				batch.Query(query, params...)
			}
			if err := m.database.GetDB().(*gocql.Session).ExecuteBatch(batch); err != nil {
				fmt.Println("Failed to use batch on no sql:", err)
			}
		}
		m.messages = m.messages[:0]
		m.messageMutex.Unlock()
	}
	queues := m.subscribers[subject]
	if (len(queues) > 0) {
		var wg sync.WaitGroup
		wg.Add(len(queues))
		for _, queue := range queues {
			eachQueue := queue
			go func(eachQueue chan<- broker.Message, msg broker.Message, m *Module) {
				wg.Done()
				select {
					case eachQueue <- msg:
				}
			}(eachQueue, msg, m)
		}
		wg.Wait()
	}
	return msg.Id, nil
}

func (m *Module) Subscribe(ctx context.Context, subject string) (<-chan broker.Message, error) {
	m.Lock()
	defer m.Unlock()

	subscriber := make(chan broker.Message, 1)
	if (m.isClosed) {
		return subscriber, broker.ErrUnavailable
	}
	m.subscribers[subject] = append(m.subscribers[subject], subscriber)
	// m.notifyPublish()

	// go func() {
	// 	<-ctx.Done()
	// 	m.Lock()
	// 	defer m.Unlock()

	// 	subscribers := m.subscribers[subject]
	// 	for i, sub := range subscribers {
	// 		if sub == subscriber {
	// 			m.subscribers[subject] = append(subscribers[:i], subscribers[i+1:]...)
	// 			prometheus.ActiveSubscribers.Dec()
	// 			break
	// 		}
	// 	}
	// 	close(subscriber)
	// }()
	return subscriber, nil
}

func (m *Module) Fetch(ctx context.Context, subject string, id int) (broker.Message, error) {
	m.RLock()
	defer m.RUnlock()
	if (m.isClosed) {
		return broker.Message{}, broker.ErrUnavailable
	}

	msg := m.findMessage(id)
	// msg := m.ramFindMessage(id)
	if msg == nil {return broker.Message{}, broker.ErrInvalidID}
	if msg.Expiration > 0 && time.Now().Before(msg.PublishDate.Add(msg.Expiration)) {
		return broker.Message{}, broker.ErrExpiredID
	}
	return *msg, nil
}

func (m *Module) findMessage(id int) (*broker.Message) {
	message := &broker.Message{}
	params := []interface{}{id}
	var expirationStr string
	var publishDate string
	if (m.database.GetType() == "sql") {
		query := "SELECT * FROM messages WHERE id = $1 LIMIT 1"
		var row *sql.Row = (m.database.SELECT(query, params...)).(*sql.Row)
		err := row.Scan(&message.Id, &message.Body, &expirationStr, &publishDate)
		if err != nil {
			if err == sql.ErrNoRows {log.Fatalf("no row found in the database -- from fetch")}
			return nil
		}
	} else {
		query := "SELECT id, body, expiration, publish_date FROM messages WHERE id = ? LIMIT 1"
		res := m.database.SELECT(query, params...).(map[string]interface{})
		expirationStr = res["expiration"].(string)
		publishDate = res["publish_date"].(string)
		message.Id = int(res["id"].(int64))
		message.Body = res["body"].(string)
	}
	message.Expiration, _ = time.ParseDuration(expirationStr)
	message.PublishDate, _ = time.Parse("2006-01-02 15:04:05.999999-07:00", publishDate)
	return message
}

func existsInSlice(message broker.Message, slice []broker.Message) int {
	for _, item := range slice {
		if item.Body == message.Body {
			return item.Id
		}
	}
	return 0
}

func (m *Module) notifyPublish()  {
	select {
		case m.checkMsg <- 2 :
			//log.Printf("a channel has been set --- publish must be notified")
		default:
			close(m.checkMsg)
	}
}
