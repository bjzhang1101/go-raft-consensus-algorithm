package server

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

// Server is the gRPC server of the Raft node.
type Server struct {
	pb.UnimplementedTickerServer

	node *node.Node
}

// NewServer returns a gRPC server.
func NewServer(node *node.Node) *grpc.Server {
	log.Printf("starting grpc server listening on :%d", DefaultPort)
	s := grpc.NewServer()

	pb.RegisterTickerServer(s, &Server{node: node})
	return s
}

// AppendEntry implemented ticker.AppendEntry.
func (s *Server) AppendEntry(ctx context.Context, in *pb.TickRequest) (*pb.TickResponse, error) {
	term := int(in.GetTerm())
	log.Printf("processing append entry request - id: %s, term: %d, data: %s", in.GetId(), term, in.GetData())

	s.node.AppendEntryC(in.GetData())

	if s.node.GetCurTerm() > term {
		return &pb.TickResponse{Accept: false}, nil
	} else {
		s.node.SetCurTerm(term)
		return &pb.TickResponse{Accept: true}, nil
	}
}

// RequestVote implemented ticker.RequestVote.
func (s *Server) RequestVote(ctx context.Context, in *pb.TickRequest) (*pb.TickResponse, error) {
	term := int(in.GetTerm())
	log.Printf("processing request vote request - id: %s, term: %d, data: %s", in.GetId(), term, in.GetData())

	s.node.AppendEntryC(in.GetData())

	if _, ok := s.node.GetVotedPerTerm()[term]; ok {
		return &pb.TickResponse{Accept: false}, nil
	} else {
		s.node.SetVotedForTerm(term)
		return &pb.TickResponse{Accept: true}, nil
	}
}
