package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	pb "therealbroker/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var 
(
	addr = flag.String("addr", "envoy-service:10000", "the address to connect to")
)

func main() {
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()),)
	if err != nil {log.Fatalf("did not connect: %v", err)}
	defer conn.Close()
	c := pb.NewBrokerClient(conn)
	mainCtx := context.Background()
	subReq := &pb.SubscribeRequest{Subject: "bootcamp",}
	res, subErr := c.Subscribe(mainCtx, subReq)
	if subErr != nil {
		log.Println("error in subscribe result")
		fmt.Println(subErr)
	}
	x, _ := res.Recv()
	fmt.Println(x)
}
