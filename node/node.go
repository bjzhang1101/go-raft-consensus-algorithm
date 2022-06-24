package node

import (
	"context"
	"math/rand"
	"strings"
	"time"
)

const (
	baseElectionTimeoutMs   = 150
	jitterElectionTimeoutMs = 150

	tickInterval = 50 * time.Millisecond
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

	// EntryC is the channel for receiving
	EntryC <-chan Tick

	// Votes is the vote number the node gets after it sends RequestVote
	// requests to other nodes for leader election.
	//
	// Only valid when the node is in Candidate state and will always start
	// with value 1 before sending RequestVote requests.
	Votes int

	// Voted indicates whether the follower votes to any Candidates.
	//
	// Should be set to false when start as a Follower.
	Voted bool

	// TickC is the channel for sending tick.
	//
	// Only valid when the node is in Leader state.
	TickC chan Tick
}

// NewNode returns a new Raft node.
func NewNode(id string, quorum []string) *Node {
	n := Node{
		ID:      id,
		State:   Follower,
		Quorum:  quorum,
		CurTerm: 0,
	}

	n.SetElectionTimeout()
	return &n
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

// SetElectionTimeout refreshes the node's election timeout by re-generate a
// random time duration.
func (n *Node) SetElectionTimeout() {
	rand.Seed(time.Now().UnixNano())
	n.ElectionTimeout = time.Duration(rand.Intn(baseElectionTimeoutMs)+jitterElectionTimeoutMs) * time.Millisecond
}

// SetVoted sets whether the node already votes to a Candidate.
func (n *Node) SetVoted(voted bool) {
	n.Voted = voted
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
	n.SetVoted(false)

	for {
		select {
		case <-time.After(n.ElectionTimeout):
			n.SetState(Candidate)
		default:
		}
	}
}

func (n *Node) startCandidate(ctx context.Context) error {
	n.SetVoted(true)
	n.SetElectionTimeout()

	votes, err := n.requestVote(ctx)
	if err != nil {
		return err
	}

	if votes > len(n.Quorum)/2 {
		n.SetState(Leader)
	} else {
		n.SetState(Follower)
	}

	return nil
}

func (n *Node) startLeader(ctx context.Context) error {
	for {
		select {
		case <-time.After(tickInterval):
			if err := n.appendEntries(ctx, Tick{}); err != nil {
				// TODO: handle error.
			}
		case entry := <-n.EntryC:
			if err := n.appendEntries(ctx, entry); err != nil {
				// TODO: handle error.
			}
		default:
		}
	}
}

func (n *Node) requestVote(ctx context.Context) (int, error) {
	votes := 1

	// TODO: update the RPC request to async.
	for _, member := range n.Quorum {
		if strings.Compare(member, n.ID) == 0 {
			continue
		}

		// resp := n.SendRequest(member)
		resp := TickResp{}

		if resp.GetConsensus() {
			votes++
		}
	}
	return votes, nil
}

func (n *Node) appendEntries(ctx context.Context, tick Tick) error {
	return nil
}
