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
	subReq3 := &pb.SubscribeRequest{Subject: "bootcamp2",}
	res3, subErr := c.Subscribe(mainCtx, subReq3)
	if subErr != nil {
		log.Println("error in subscribe result")
		fmt.Println(subErr)
	}
	x3, _ := res3.Recv()
	fmt.Println(x3)
}
