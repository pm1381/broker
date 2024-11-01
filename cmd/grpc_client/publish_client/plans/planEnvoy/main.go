package main

import (
	"context"
	"log"
	"time"
	pb "therealbroker/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ticker := time.NewTicker(20 * time.Millisecond)
	conn, err := grpc.Dial("envoy-service:10000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {log.Fatalf("failed to dial: ", err)}
	defer conn.Close()
	client := pb.NewBrokerClient(conn)
	ctx := context.Background()
	for {
		select {
		case <-context.Background().Done():
			return
		case <-ticker.C:
			go func (client pb.BrokerClient, ctx context.Context)  {
				go PublishNewConnection(client, ctx)
			}(client, ctx)
		}
	}
}

func PublishNewConnection(client pb.BrokerClient, ctx context.Context) {
	pubReq := &pb.PublishRequest{
		Subject: "bale.bootcamp",
		Body: []byte("test"),
		ExpirationSeconds: 0,
	}
	_, err := client.Publish(ctx, pubReq)
	if err != nil {log.Printf("Publish error: ", err)}
}
