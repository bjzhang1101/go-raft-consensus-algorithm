package grpc

import (
	"context"
	"log"
	"math"

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

// AppendEntries implemented ticker.AppendEntries.
//
// If the leader's current term is smaller than the node's, reject the request
// and return its current term.  If the follower does not find an entry in its
// log with the same index and term, then it refuses the new entries.
func (s *Server) AppendEntries(ctx context.Context, in *pb.TickRequest) (*pb.TickResponse, error) {
	leaderCurTerm := int(in.GetLeaderCurTerm())
	curTerm := s.node.GetCurTerm()

	// Term check.
	if curTerm > leaderCurTerm {
		log.Printf("current leader %s term is stale, reject the request", in.GetLeaderId())
		return &pb.TickResponse{Accept: false, Term: int32(curTerm)}, nil
	}

	s.node.AppendTickC()
	s.node.SetCurTerm(leaderCurTerm)
	s.node.SetCurLeader(in.GetLeaderId())

	// Return directly for heartbeat.
	if len(in.GetEntries()) == 0 {
		s.node.SetCommitIdx(int(math.Min(float64(in.GetLeaderCommitIdx()), float64(len(s.node.GetLogs())-1))))
		return &pb.TickResponse{Accept: true, Term: int32(curTerm)}, nil
	}

	log.Printf("processing append entry request - id: %s, term: %d", in.GetLeaderId(), leaderCurTerm)

	// Consistency check.
	logs := s.node.GetLogs()
	prevLogIdx := int(in.GetPrevLogIdx())

	if len(logs) <= prevLogIdx {
		return &pb.TickResponse{Accept: false, Term: int32(curTerm)}, nil
	}

	if logs[prevLogIdx].GetTerm() != in.GetPrevLogTerm() {
		return &pb.TickResponse{Accept: false, Term: int32(curTerm)}, nil
	}

	// Override entries.
	newLogs := logs[:prevLogIdx+1]
	for _, e := range in.GetEntries() {
		newLogs = append(newLogs, e)
	}

	s.node.SetLogs(newLogs)
	s.node.SetCommitIdx(int(math.Min(float64(in.GetLeaderCommitIdx()), float64(len(s.node.GetLogs())-1))))
	return &pb.TickResponse{Accept: true, Term: int32(curTerm)}, nil
}

// RequestVote implemented ticker.RequestVote.
func (s *Server) RequestVote(ctx context.Context, in *pb.TickRequest) (*pb.TickResponse, error) {
	term := int(in.GetLeaderCurTerm())
	log.Printf("processing request vote request - id: %s, term: %d", in.GetLeaderId(), term)

	// If a Follower has the same term with Candidate, this most likely means
	// they are all requesting votes for leader election.
	if s.node.GetCurTerm() >= term {
		return &pb.TickResponse{Accept: false, Term: int32(s.node.GetCurTerm())}, nil
	}

	if _, ok := s.node.GetVotedPerTerm()[term]; ok {
		return &pb.TickResponse{Accept: false, Term: int32(s.node.GetCurTerm())}, nil
	}

	s.node.AppendTickC()
	s.node.SetVotedForTerm(term)
	return &pb.TickResponse{Accept: true, Term: int32(s.node.GetCurTerm())}, nil
}
