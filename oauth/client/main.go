package main

import (
	"context"
	"crypto/tls"
	"log"
	"time"

	pb "github.com/ryutah/go-grpc-sample/helloworld/helloworld/go"

	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

func main() {
	perRPC := oauth.NewOauthAccess(fetchToken())
	opts := []grpc.DialOption{
		grpc.WithPerRPCCredentials(perRPC),
		grpc.WithTransportCredentials(
			credentials.NewTLS(&tls.Config{InsecureSkipVerify: true}),
		),
	}
	conn, err := grpc.Dial(":8080", opts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.SayHello(ctx, &pb.HelloRequest{
		Name: "Auth sample client",
	})
	if err != nil {
		log.Fatalf("could not greet: %#v", err)
	}
	log.Printf("Greeting: %v", r.Message)
}

func fetchToken() *oauth2.Token {
	return &oauth2.Token{
		AccessToken: "some-secret-token",
	}
}
