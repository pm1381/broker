package db

import (
	_ "github.com/lib/pq"
)


type Database interface {
	Connect() error
	Disconnect() error
	GetDB() interface{}
	GetType() string
	GetName() string
	Execute(query string, params ...interface{}) (interface{} , error) // update delete
	SELECT(query string, params ...interface{}) (interface{})
	INSERTwithLastId(query string, params ...interface{}) (interface{} , error)
}

type DBProvider struct {
	DatabaseType Database
}

func  (dp *DBProvider) Provide() (Database, error)  {
	return dp.DatabaseType, nil
}
