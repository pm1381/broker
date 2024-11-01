package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	pb "therealbroker/api/proto"
	internalBroker "therealbroker/internal/broker"
	"therealbroker/internal/router"
	broker "therealbroker/pkg/broker"
	"therealbroker/pkg/cache"
	"therealbroker/pkg/prometheus"
	"time"

	//"therealbroker/pkg/tracing"
	//"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	//"go.opentelemetry.io/otel"
	// "go.opentelemetry.io/otel/attribute"
	// "go.opentelemetry.io/otel/trace"
	"github.com/gomodule/redigo/redis"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	pb.UnimplementedBrokerServer
	newBroker broker.Broker
	cache *cache.Redis
	mu sync.Mutex
}

var (
	port = flag.Int("port", 50051, "The server port")
	httpPort = 80
)

func (s *Server) Publish(ctx context.Context, in *pb.PublishRequest) (*pb.PublishResponse, error) {
	podIp := os.Getenv("MY_POD_IP")
	prometheus.MethodCount.WithLabelValues("Publish", "success", podIp).Inc()
	startTime := time.Now()
	defer func ()  {
		duration := time.Since(startTime).Seconds()
		prometheus.MethodDuration.WithLabelValues("Publish", podIp).Observe(duration)
	}()
	log.Println("publishing... " + "pod ip: " + podIp + " on subject: " + in.GetSubject())
	message := broker.Message{
		Body: string(in.GetBody()),
		Expiration: time.Duration(in.GetExpirationSeconds()),
	}
	messageId, err := s.newBroker.Publish(ctx, in.GetSubject(), message)
	if err != nil {
		return nil, err
	}
	response := &pb.PublishResponse{
		Id: int32(messageId),
	}
	s.mu.Lock()
	exists, _ := redis.Bool(s.cache.GetConnection().Do("EXISTS", in.GetSubject()))
	if exists {
		s.cache.GetConnection().Do("SREM", in.GetSubject(), "test")
	}
	log.Println("message published"  + " for subject: " + in.GetSubject())
	s.mu.Unlock()
	return response, nil
}

func (s *Server) Subscribe(req *pb.SubscribeRequest, stream pb.Broker_SubscribeServer) error {
	// _, span := tracer.Start(stream.Context(), "Subscribe", trace.WithAttributes(attribute.String("subscribe", "server")))
	// defer span.End()
	podIp := os.Getenv("MY_POD_IP")
	prometheus.MethodCount.WithLabelValues("Subscribe", "success", podIp).Inc()
	startTime := time.Now()
	defer func ()  {
		duration := time.Since(startTime).Seconds()
		prometheus.MethodDuration.WithLabelValues("Subscribe", podIp).Observe(duration)
	}()
	s.cache.GetConnection().Do("SADD", req.GetSubject(), podIp)
	log.Println("subscribing... " + "pod ip: " + podIp + " on subject: " + req.GetSubject())

	newChannel, err := s.newBroker.Subscribe(stream.Context(), req.GetSubject())
	if err != nil {return err}
	for msg := range newChannel {
		messageRepsonse := pb.MessageResponse{Body:  []byte(msg.Body)}
		err := stream.Send(&messageRepsonse)
		if err != nil {return err}
	}
	return nil
}

func (s *Server) Fetch(ctx context.Context, req *pb.FetchRequest) (*pb.MessageResponse, error) {
	//log.Printf("start fetching")
	// _, span := tracer.Start(ctx, "Fetch", trace.WithAttributes(attribute.String("fetch", "server")))
	// defer span.End()
	msg, err := s.newBroker.Fetch(ctx, req.GetSubject(), int(req.GetId()))
	if err != nil {
		return nil, err
	}
	return &pb.MessageResponse{Body:   []byte(msg.Body)}, nil
}

func publishMiddleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error)  {
	canProceed := false
	if info.FullMethod == "/broker.Broker/Publish" {
		subject := req.(*pb.PublishRequest).Subject
		currentPod := os.Getenv("MY_POD_IP")
		info.Server.(*Server).mu.Lock()
		exists, _ := redis.Bool(info.Server.(*Server).cache.GetConnection().Do("EXISTS", subject))
		if exists {
			targetPods, err := redis.Strings(info.Server.(*Server).cache.GetConnection().Do("SMEMBERS", subject))
			if currentPod != "" && err == nil {
				for _, targetPod := range targetPods {
					if (currentPod == targetPod) {
						log.Println("target pod and current pod are the same")
						handler(ctx, req)
					} else if (targetPod != "") {
						log.Println("target pod and current pod are different... changing dest. to new pod ip")
						conn, err := grpc.Dial(targetPod + ":50051", grpc.WithTransportCredentials(insecure.NewCredentials()),)
						if err != nil {log.Fatalf("did not connect: %v", err)}
						defer conn.Close()
						c := pb.NewBrokerClient(conn)
						c.Publish(ctx, req.(*pb.PublishRequest))
					}
				}
			} else if (err == nil && currentPod == "") {
				canProceed = true
			} else {
				log.Println("error occured while reading from redis")
				fmt.Println(err)
			}
		} else {
			canProceed = true
		}
		info.Server.(*Server).mu.Unlock()
	} else {
		canProceed = true
	}
	if canProceed {return handler(ctx, req)}
	return nil, nil
}

func main()  {
	startGrpcServer()
	startHttpServer()
	for {
	}
}

func startHttpServer()  {
	router.AllRoutes()
	go func() {
		log.Printf("HTTP server listening at :%d", httpPort)
		if err := http.ListenAndServe(":80", nil); err != nil {
			log.Fatalf("failed to serve HTTP: %v", err)
		}
	}()
}

func startGrpcServer()  {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {log.Fatalf("failed to listen: %v", err)}
	s := grpc.NewServer(grpc.UnaryInterceptor(publishMiddleware))
	redis := cache.NewRedis("Redis")
	redis.Connect()
	pb.RegisterBrokerServer(s, &Server{newBroker: internalBroker.NewModule(), cache: redis})
	go func() {
		log.Printf("gRPC server listening at %v", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()
}