package node

import (
	"context"
	"log"
	"math/rand"
	"strings"
	"time"

	grpc "github.com/bjzhang1101/raft/grpc/client"
)

const (
	baseElectionTimeoutMs   = 5000
	jitterElectionTimeoutMs = 5000

	tickInterval = 2500 * time.Millisecond
)

// Node includes all the metadata of a node.
type Node struct {
	// ID is the ID of the node.
	ID string

	// State is the state of the node: Follower, Candidate or Leader.
	State State

	// Quorum is a list of IP addresses that includes all the nodes that
	// forms the Raft cluster.
	Quorum []string

	// TickClients is a list of clients for nodes in the Raft cluster.
	TickClients []*grpc.Client

	// CurTerm is the current term of the node.
	//
	// If one server’s current term is smaller than the other’s, then it
	// updates its current term to the larger value. If it receives a request
	// with a stale term number, it rejects the request.
	// If a candidate or leader discovers that its term is out of date, it
	// immediately reverts to follower state.
	CurTerm int

	// ElectionTimeout is the timeout for a follower or a candidate to start
	// (or restart for Candidate) leader election if it receives no
	// communication over the period of time.
	ElectionTimeout time.Duration

	// EntryC is the channel for received entries.
	EntryC chan string

	// VotedPerTerm indicates whether the follower votes to any Candidate in
	// a specific term.
	VotedPerTerm map[int]struct{}
}

// NewNode returns a new Raft node.
func NewNode(id string, quorum []string, grpcClientPort int) *Node {
	log.Printf("initializing raft node %s", id)
	n := Node{
		ID:      id,
		State:   Follower,
		Quorum:  quorum,
		CurTerm: 0,
	}

	var tickClients []*grpc.Client
	for _, member := range quorum {
		// Skip creating gRPC client for itself.
		if strings.Compare(member, id) == 0 {
			continue
		}

		tickClients = append(tickClients, grpc.NewClient(member, grpcClientPort))
	}

	n.TickClients = tickClients

	rand.Seed(time.Now().UnixNano())

	n.SetElectionTimeout()

	n.VotedPerTerm = make(map[int]struct{})
	n.EntryC = make(chan string)
	return &n
}

// GetID returns the node's ID.
func (n *Node) GetID() string {
	return n.ID
}

// GetState returns the node's current state.
func (n *Node) GetState() State {
	return n.State
}

// SetState sets the node's current state.
func (n *Node) SetState(state State) {
	n.State = state
}

// GetCurTerm returns the node's current term.
func (n *Node) GetCurTerm() int {
	return n.CurTerm
}

// IncrementCurTerm increase 1 to the current term.
func (n *Node) IncrementCurTerm() {
	n.CurTerm++
}

// SetCurTerm sets the current term.
func (n *Node) SetCurTerm(term int) {
	n.CurTerm = term
}

// SetElectionTimeout refreshes the node's election timeout by re-generate a
// random time duration.
func (n *Node) SetElectionTimeout() {
	n.ElectionTimeout = time.Duration(rand.Intn(baseElectionTimeoutMs)+jitterElectionTimeoutMs) * time.Millisecond
}

// GetVotedPerTerm returns the hash set of voted term.
func (n *Node) GetVotedPerTerm() map[int]struct{} {
	return n.VotedPerTerm
}

// SetVotedForTerm adds the term in the VotedPerTerm hash set.
func (n *Node) SetVotedForTerm(term int) {
	n.VotedPerTerm[term] = struct{}{}
}

// AppendEntryC appends the entry channel with the data.
func (n *Node) AppendEntryC(data string) {
	n.EntryC <- data
}

// Start is the main goroutine for a node's main functionality.
func (n *Node) Start(ctx context.Context) <-chan struct{} {
	done := make(chan struct{})
	var err error

	for {
		switch n.State {
		case Follower:
			err = n.startFollower(ctx)
		case Candidate:
			err = n.startCandidate(ctx)
		case Leader:
			err = n.startLeader(ctx)
		}

		if err != nil {
			close(done)
			return done
		}
	}
}

func (n *Node) startFollower(ctx context.Context) error {
	log.Printf("node %s became Follower", n.ID)
	for {
		select {
		case <-time.After(n.ElectionTimeout):
			n.SetState(Candidate)
			return nil
		case data := <-n.EntryC:
			n.SetElectionTimeout()
			log.Printf("node %s received data %s", n.ID, data)
		}
	}
}

func (n *Node) startCandidate(ctx context.Context) error {
	log.Printf("node %s became Candidate", n.ID)

	n.SetElectionTimeout()
	n.IncrementCurTerm()
	n.SetVotedForTerm(n.GetCurTerm())

	votes, err := n.requestVote(ctx)
	if err != nil {
		return err
	}

	log.Printf("node %s receives %d votes in term %d", n.ID, votes, n.GetCurTerm())

	if votes > len(n.Quorum)/2 {
		n.SetState(Leader)
	} else {
		n.SetState(Follower)
	}

	return nil
}

func (n *Node) startLeader(ctx context.Context) error {
	log.Printf("node %s became Leader", n.ID)

	for {
		select {
		case <-time.After(tickInterval):
			if err := n.appendEntries(ctx, "this is a heartbeat"); err != nil {
				log.Fatalf("node %s failed to send tick", n.ID)
			}
		}
	}
}

func (n *Node) requestVote(ctx context.Context) (int, error) {
	votes := 1

	for _, c := range n.TickClients {
		accepted := c.RequestVote(ctx, n.GetID(), n.GetCurTerm(), "this is a request vote request")
		log.Printf("node %s votes for %v", c.GetAddress(), accepted)

		if accepted {
			votes++
		}
	}

	return votes, nil
}

func (n *Node) appendEntries(ctx context.Context, data string) error {
	for _, c := range n.TickClients {
		resp := c.AppendEntry(ctx, n.GetID(), n.GetCurTerm(), data)
		log.Printf("receive heartbeat response %s from %s in term %d", resp, c.GetAddress(), n.GetCurTerm())
	}

	return nil
}
