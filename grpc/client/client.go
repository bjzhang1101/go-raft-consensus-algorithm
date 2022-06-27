package client

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/bjzhang1101/raft/grpc/protobuf"
)

// Client is a gRPC client for a Raft node.
type Client struct {
	c pb.TickerClient

	address string
	port    int
}

// GetAddress returns the destination address of the client.
func (c *Client) GetAddress() string {
	return c.address
}

// NewClient returns a new gRPC client sending ticks between Raft nodes.
func NewClient(a string, p int) *Client {
	log.Printf("starting new gRPC client to %s:%d", a, p)
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", a, p), grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	// TODO: investigate how to close the connection properly.
	// defer conn.Close()

	gc := pb.NewTickerClient(conn)

	c := Client{c: gc, address: a, port: p}

	return &c
}

// AppendEntry is the function that the Leader sent to followers to sync
// data and keep leadership.
func (c *Client) AppendEntry(ctx context.Context, id string, term int, data string) (bool, error) {
	log.Println("sending append entry request")
	r, err := c.c.AppendEntry(ctx, &pb.TickRequest{
		Id:   id,
		Term: int32(term),
		Data: data,
	})

	if err != nil {
		return true, fmt.Errorf("failed to append entry: %v", err)
	}

	return r.GetAccept(), nil
}

// RequestVote is the function that the Candidate sent to followers to
// requests their votes for leader election.
func (c *Client) RequestVote(ctx context.Context, id string, term int, data string) (bool, error) {
	log.Printf("node %s sending request vote request", id)
	r, err := c.c.RequestVote(ctx, &pb.TickRequest{
		Id:   id,
		Term: int32(term),
		Data: data,
	})

	if err != nil {
		return false, fmt.Errorf("failed to request vote: %v", err)
	}

	return r.GetAccept(), nil
}
