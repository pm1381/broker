package dbtype

import (
	"log"

	"github.com/gocql/gocql"
)

type Scylla struct {
	cluster *gocql.ClusterConfig
	session *gocql.Session
	Name string
	Type string
}

func (c *Scylla) Connect() error {
	// k8s : scylla-service:9042
	// others : localhost:9040
	if c.cluster == nil && c.session == nil {
		cluster := gocql.NewCluster("scylla-service:9042")
		cluster.Keyspace = "broker"
		cluster.Consistency = gocql.Quorum
		session, err := cluster.CreateSession()
		if err != nil {
			return err
		}
		log.Println("connected to scylla on port 9042")
		c.cluster = cluster
		c.session = session
	}
	return nil
}

func (c *Scylla) Disconnect() error {
	c.session.Close()
	return nil
}


func (c *Scylla) INSERTwithLastId(query string, params ...interface{}) (interface{}, error) {
	return nil, nil
}

func (c *Scylla) Execute(query string, params ...interface{}) (interface{}, error) {
	res := c.session.Query(query, params...).Exec()
	// fmt.Println(res)
	return res, nil
}

func (c *Scylla) GetDB() (interface{}) {
	return c.session
}

func (c *Scylla) GetType() (string) {
	return c.Type
}

func (c *Scylla) GetName() (string) {
	return c.Name
}

func (c *Scylla) SELECT(query string, params ...interface{}) (interface{}) {
	//must be implemented
	iter := c.session.Query(query, params...).Iter()
	defer iter.Close()
	row := make(map[string]interface{})
	for {
		if !iter.MapScan(row) {
			break
		}	
	}
	return row
}
