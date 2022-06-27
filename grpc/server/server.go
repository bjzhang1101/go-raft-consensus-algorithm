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

	if s.node.GetCurTerm() > term {
		return &pb.TickResponse{Accept: false}, nil
	}

	if len(in.GetData()) > 0 {
		s.node.AppendEntryC(in.GetData())
	} else {
		s.node.AppendTickC()
	}

	s.node.SetCurTerm(term)
	return &pb.TickResponse{Accept: true}, nil
}

// RequestVote implemented ticker.RequestVote.
func (s *Server) RequestVote(ctx context.Context, in *pb.TickRequest) (*pb.TickResponse, error) {
	term := int(in.GetTerm())
	log.Printf("processing request vote request - id: %s, term: %d, data: %s", in.GetId(), term, in.GetData())

	// If a Follower has the same term with Candidate, this most likely means
	// they are all requesting votes for leader election.
	if s.node.GetCurTerm() >= term {
		return &pb.TickResponse{Accept: false}, nil
	}

	if _, ok := s.node.GetVotedPerTerm()[term]; ok {
		return &pb.TickResponse{Accept: false}, nil
	}

	s.node.SetVotedForTerm(term)
	s.node.AppendTickC()
	s.node.SetCurTerm(term)
	return &pb.TickResponse{Accept: true}, nil
}
