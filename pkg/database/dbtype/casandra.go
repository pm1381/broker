package dbtype

import (

	"github.com/gocql/gocql"
)

type CassandraDatabase struct {
	cluster *gocql.ClusterConfig
	session *gocql.Session
	Name string
	Type string
}

func (c *CassandraDatabase) Connect() error {
	if c.cluster == nil && c.session == nil {
		cluster := gocql.NewCluster("127.0.0.1:9042")
		cluster.Keyspace = "broker"
		cluster.Consistency = gocql.Quorum
		session, err := cluster.CreateSession()
		if err != nil {
			return err
		}
		c.cluster = cluster
		c.session = session
	}
	return nil
}

func (c *CassandraDatabase) Disconnect() error {
	c.session.Close()
	return nil
}


func (c *CassandraDatabase) INSERTwithLastId(query string, params ...interface{}) (interface{}, error) {
	return nil, nil
}

func (c *CassandraDatabase) Execute(query string, params ...interface{}) (interface{}, error) {
	res := c.session.Query(query, params...).Exec()
	// fmt.Println(res)
	return res, nil
}

func (c *CassandraDatabase) GetDB() (interface{}) {
	return c.session
}

func (c *CassandraDatabase) GetType() (string) {
	return c.Type
}

func (c *CassandraDatabase) GetName() (string) {
	return c.Name
}

func (c *CassandraDatabase) SELECT(query string, params ...interface{}) (interface{}) {
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
