package grpc

import (
	"context"
	"log"

	"google.golang.org/grpc"

	pb "github.com/bjzhang1101/raft/grpc/protobuf"
	"github.com/bjzhang1101/raft/node"
)

const (
	// DefaultPort is the default gRPC port of a node.
	DefaultPort = 8081
)

// RaftServer is the gRPC server of the Raft node.
type RaftServer struct {
	pb.UnimplementedGreeterServer

	node *node.Node
}

// NewServer returns a gRPC server.
func NewServer(node *node.Node) *grpc.Server {
	s := grpc.NewServer()

	pb.RegisterGreeterServer(s, &RaftServer{node: node})
	return s
}

// SayHello implements helloworld.GreeterServer
func (s *RaftServer) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName() + "state " + s.node.GetState().String()}, nil
}
