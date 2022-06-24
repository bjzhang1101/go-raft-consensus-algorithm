package grpc

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/bjzhang1101/raft/grpc/protobuf"
)

const (
	defaultName = "world"
)

type RaftClient struct {
	Client pb.GreeterClient

	Address string
	Port    int
}

func NewClient(a string, p int) *RaftClient {
	// Set up a connection to the server.
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", a, p), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	gc := pb.NewGreeterClient(conn)

	c := RaftClient{
		Client: gc,
		Port:   8081,
	}

	return &c
}

func (c *RaftClient) SayHello(ctx context.Context) string {
	r, err := c.Client.SayHello(ctx, &pb.HelloRequest{Name: defaultName})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	return r.GetMessage()
}
