//go:generate protoc -I ./helloworld --java_out ./helloworld/java --js_out ./helloworld/js --python_out ./helloworld/python --go_out=plugins=grpc:./helloworld/go ./helloworld/helloworld.proto

package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "github.com/ryutah/go-grpc-sample/helloworld/go"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type greeterServer struct{}

func (g *greeterServer) SayHello(ctx context.Context, r *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Get Message: %#v", r)

	return &pb.HelloReply{
		Message: fmt.Sprintf("Hello, %s!", r.Name),
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen port :8080; %#v", err)
	}

	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, new(greeterServer))
	reflection.Register(s)

	log.Println("Start server on port :8080...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to start server %#v", err)
	}
}
