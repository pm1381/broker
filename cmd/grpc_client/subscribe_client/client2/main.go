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
	addr = flag.String("addr", "broker-service:50051", "the address to connect to")
)

func main() {
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()),)
	if err != nil {log.Fatalf("did not connect: %v", err)}
	defer conn.Close()
	c := pb.NewBrokerClient(conn)
	mainCtx := context.Background()
	subReq2 := &pb.SubscribeRequest{Subject: "bootcamp1",}
	res2, subErr := c.Subscribe(mainCtx, subReq2)
	if subErr != nil {
		log.Println("error in subscribe result")
		fmt.Println(subErr)
	}
	x2, _ := res2.Recv()
	fmt.Println(x2)
}
