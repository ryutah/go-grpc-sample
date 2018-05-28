package main

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"strings"

	pb "github.com/ryutah/go-grpc-sample/helloworld/helloworld/go"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/testdata"
)

type greeterServer struct{}

func (g *greeterServer) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Get Message: %#v", in)
	return &pb.HelloReply{
		Message: "Hello " + in.Name,
	}, nil
}

func main() {
	log.Println("server starting on port 8080...")
	cert, err := tls.LoadX509KeyPair(testdata.Path("server1.pem"), testdata.Path("server1.key"))
	if err != nil {
		log.Fatalf("failed to load key pair: %#v", err)
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(ensureValidToken),
		grpc.Creds(credentials.NewServerTLSFromCert(&cert)),
	}

	s := grpc.NewServer(opts...)
	pb.RegisterGreeterServer(s, new(greeterServer))
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listne %#v", err)
	}

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve %#v", err)
	}
}

var (
	errMissingMetadata = status.Errorf(codes.InvalidArgument, "missing metadata")
	errInvalidToken    = status.Errorf(codes.Unauthenticated, "invalid token")
)

func ensureValidToken(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	valid := func(authorization []string) bool {
		if len(authorization) < 1 {
			return false
		}
		token := strings.TrimPrefix(authorization[0], "Bearer ")
		return token == "some-secret-token"
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errMissingMetadata
	}
	if !valid(md["authorization"]) {
		return nil, errInvalidToken
	}
	return handler(ctx, req)
}
