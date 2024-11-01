package cache

import (
	"log"

	"github.com/gomodule/redigo/redis"
)

type Redis struct {
	connection redis.Conn
	name string
}

func NewRedis(name string) *Redis {
	return &Redis{
		name: name,
	}
}

func (r *Redis) Connect() {
	if r.connection == nil {
		//redis docker: localhost:6379
		//redis kuber: redis-service:6379
		conn, err := redis.Dial("tcp", "localhost:6379")
		if err != nil {
			log.Fatalf("cannot connect to redis server")
		}
		log.Println("connected to redis on port 6379")
		r.connection = conn
	}
}

func (r *Redis) Disconnect() {
	r.connection.Close()
}

func (r *Redis) GetConnection() redis.Conn  {
	return r.connection
}

func (r *Redis) GetName() string {
	return r.name
}