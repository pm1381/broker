package dbtype

import (
	"database/sql"
	"log"
	//"fmt"

	_ "github.com/lib/pq"
)

type Pgsql struct {
	db *sql.DB
	Name string
	Type string
}

func (p *Pgsql) Connect() error {
	if (p.db == nil) {
		// connStr := "postgres://parham:ParhamBootcamp8102@localhost:5432/broker?sslmode=disable"
		localStr := "user=parham password=1381pm dbname=broker sslmode=disable"
		// k8sStr := "postgres://parham:ParhamBootcamp8102@psql-service:5432/broker?sslmode=disable"
		db, err := sql.Open("postgres", localStr)
		if err != nil {
			return err
		}
		log.Println("connected to posgress on port 5432")
		p.db = db
	}
	return nil
}

func (p *Pgsql) Disconnect() error {
	return p.db.Close()
}

func (p *Pgsql) Execute(query string, params ...interface{}) (interface{}, error) {
	res, err := p.db.Exec(query, params...)
	return res, err
}

func (p *Pgsql) INSERTwithLastId(query string, params ...interface{}) (interface{}, error) {
	query = query + " RETURNING id "
	res := p.db.QueryRow(query, params...)
	return res, nil
}

func (p *Pgsql) GetDB() (interface{}) {
	return p.db
}

func (p *Pgsql) GetType() (string) {
	return p.Type
}

func (p *Pgsql) GetName() (string) {
	return p.Name
}

func (p *Pgsql) SELECT(query string, params ...interface{}) (interface{}) {
	row := p.db.QueryRow(query, params...)
	return row
}
