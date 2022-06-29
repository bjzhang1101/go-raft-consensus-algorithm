package grpc

import (
	"context"
	"log"

	"google.golang.org/grpc"

	"github.com/bjzhang1101/raft/node"
	pb "github.com/bjzhang1101/raft/protobuf"
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
	entry := in.GetEntry()
	term := int(entry.GetTerm())
	log.Printf("processing append entry request - id: %s, term: %d, action: %s", in.GetId(), term, entry.GetAction())

	if s.node.GetCurTerm() > term {
		return &pb.TickResponse{Accept: false, Term: int32(s.node.GetCurTerm())}, nil
	}

	if in.GetEntry().GetAction() == pb.Entry_Tick {
		s.node.AppendTickC()
	} else {
		s.node.AppendEntryC(node.NewEntry(entry.Action, entry.GetKey(), entry.GetValue(), term))
	}

	s.node.SetCurTerm(term)
	s.node.SetCurLeader(in.GetId())
	return &pb.TickResponse{Accept: true, Term: int32(s.node.GetCurTerm())}, nil
}

// RequestVote implemented ticker.RequestVote.
func (s *Server) RequestVote(ctx context.Context, in *pb.TickRequest) (*pb.TickResponse, error) {
	term := int(in.GetEntry().GetTerm())
	log.Printf("processing request vote request - id: %s, term: %d, action: %s", in.GetId(), term, in.GetEntry().GetAction())

	// If a Follower has the same term with Candidate, this most likely means
	// they are all requesting votes for leader election.
	if s.node.GetCurTerm() >= term {
		return &pb.TickResponse{Accept: false, Term: int32(s.node.GetCurTerm())}, nil
	}

	if _, ok := s.node.GetVotedPerTerm()[term]; ok {
		return &pb.TickResponse{Accept: false, Term: int32(s.node.GetCurTerm())}, nil
	}

	s.node.SetVotedForTerm(term)
	s.node.AppendTickC()
	return &pb.TickResponse{Accept: true, Term: int32(s.node.GetCurTerm())}, nil
}
