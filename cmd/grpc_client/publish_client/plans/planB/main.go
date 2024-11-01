package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	pb "therealbroker/api/proto"
	"time"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var 
(
	addr = flag.String("addr", "broker-service:50051", "the address to connect to")
	subjects = []string{"bootcamp", "bootcamp1", "bootcamp2"}
)

func main() {
	mainCtx := context.Background()
	ticker := time.NewTicker(time.Microsecond)
	for {
		select {
		case <-ticker.C:
			conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()),)
			if err != nil {log.Fatalf("did not connect: %v", err)}
			c := pb.NewBrokerClient(conn)
			pubReq := &pb.PublishRequest{
				Subject: subjects[randomKey(3)],
				Body: []byte(randomString(10)),
				ExpirationSeconds: 0,
			}
			_, pubErr := c.Publish(mainCtx, pubReq)
			if pubErr != nil {
				fmt.Println(pubErr)
				log.Println("error in server publish")
			}
		}
	}
}

func randomString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func randomKey(n int) int {
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(n)
	return randomNumber
}