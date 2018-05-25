package main

import (
	"context"
	"log"
	"os"
	"time"

	pb "github.com/ryutah/go-grpc-sample/helloworld/go"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to server; %#v", err)
	}
	defer conn.Close()

	cli := pb.NewGreeterClient(conn)

	var name string
	if len(os.Args) > 1 {
		name = os.Args[1]
	} else {
		name = "Sample name"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := pb.HelloRequest{
		Name: name,
	}
	rep, err := cli.SayHello(ctx, &req)
	if err != nil {
		log.Fatalf("failed to receive response; %#v", err)
	}

	log.Printf("Get resposne: %#v", rep)
}
