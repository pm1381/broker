package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"
	pb "therealbroker/api/proto"
	"time"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var(
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
)

func main() {
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(
		insecure.NewCredentials()),
		// grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		// grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewBrokerClient(conn)
	mainCtx := context.Background()
	var wg sync.WaitGroup
	wg.Add(1)
	done := make(chan bool, 1)
	go func() {
		defer wg.Done()
		publishClientTest(mainCtx, c, done)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Minute)
		done <- true
	}()
	wg.Wait()
}

func publishClientTest(ctx context.Context, client pb.BrokerClient, done chan bool) {
	var wg sync.WaitGroup
	ticker := time.NewTicker(20 * time.Microsecond) // 20k msg/s 1,2M msg/min 24M msg in 20 min
	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			wg.Wait()
			return
		case <-ticker.C:
			wg.Add(1)
			go func() {
				defer wg.Done()
				pubReq := &pb.PublishRequest{
					Subject: "bale.bootcamp",
					Body: []byte("test"),
					ExpirationSeconds: 0,
				}
				_, pubErr := client.Publish(ctx, pubReq)
				if pubErr != nil {
					fmt.Println(pubErr)
					log.Println("error in server publish")
				}
			}()
		}
	}
}
