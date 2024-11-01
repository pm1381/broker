package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	pb "therealbroker/api/proto"
	"time"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	//"therealbroker/pkg/tracing"
	//"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
	// grpc-server 	  localhost:50051
	// envoy 		  127.0.0.1:10000
	// k8s-envoy 	  envoy-service:10000
	// k8s-broker     broker-service:50051
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
	pubReq := &pb.PublishRequest{
		Subject: "bale.bootcamp",
		Body: []byte("test"),
		ExpirationSeconds: 0,
	}

	ticker := time.NewTicker(time.Microsecond)
	for {
		select {
		case <-ticker.C:
			_, pubErr := c.Publish(mainCtx, pubReq)
				if pubErr != nil {
					fmt.Println(pubErr)
					log.Println("error in server publish")
				}	
		}
	}
	// subReq := &pb.SubscribeRequest{
	// 	Subject: "bale.bootcamp",
	// }
	// res, subErr := c.Subscribe(mainCtx, subReq)
	// x, _ := res.Recv()
	// fmt.Println(x)
}
