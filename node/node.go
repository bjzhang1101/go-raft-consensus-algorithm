package node

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/bjzhang1101/raft/client/grpc"
)

const (
	baseElectionTimeoutMs   = 5000
	jitterElectionTimeoutMs = 5000

	tickInterval = 2500 * time.Millisecond

	grpcReqTimeout = 500 * time.Millisecond
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

	// Log is the list of all stored log entries.
	Logs []Entry

	// Data is the node's data store.
	//
	// It could be a real data store, we use in memory map for simplicity.
	Data map[string]string

	// VotedPerTerm indicates whether the follower votes to any Candidate in
	// a specific term.
	VotedPerTerm map[int]struct{}

	// CommitIdx is the index of the highest log known to be committed.
	CommitIdx int

	// LastApplied is the index of the highest log known to be applied.
	LastApplied int

	// FollowerAttr includes attributes for follower state.
	FollowerAttr *FollowerAttributes

	// LeaderAttr includes attributes for leader state.
	LeaderAttr *LeaderAttributes
}

type FollowerAttributes struct {
	// CurLeader is the current leader of the quorum.
	CurLeader string

	// ElectionTimeout is the timeout for a follower or a candidate to start
	// (or restart for Candidate) leader election if it receives no
	// communication over the period of time.
	ElectionTimeout time.Duration

	// EntryC is the channel for received entries.
	EntryC chan Entry

	// TickC is the channel for empty entries.
	TickC chan struct{}
}

type LeaderAttributes struct {
	// TickClients is a list of clients for nodes in the Raft cluster.
	TickClients []*grpc.Client

	// NextIdx is the map of the index of the next entry to send to Followers.
	NextIdx map[string]int

	// MatchIdx is the map of the index of the highest log entry known to be
	// replicated by Followers.
	MatchIdx map[string]int
}

// NewNode returns a new Raft node.
func NewNode(id string, quorum []string, grpcClientPort int) *Node {
	log.Printf("initializing raft node %s", id)

	var logs []Entry
	data := make(map[string]string)
	n := Node{
		ID:      id,
		State:   Follower,
		Quorum:  quorum,
		CurTerm: 0,
		Logs:    logs,
		Data:    data,
	}

	n.VotedPerTerm = make(map[int]struct{})

	entryC := make(chan Entry)
	tickC := make(chan struct{})

	fa := FollowerAttributes{
		EntryC: entryC,
		TickC:  tickC,
	}

	n.FollowerAttr = &fa

	rand.Seed(time.Now().UnixNano())
	n.SetElectionTimeout()

	var tickClients []*grpc.Client
	for _, member := range quorum {
		// Skip creating gRPC client for itself.
		if strings.Compare(member, id) == 0 {
			continue
		}

		tickClients = append(tickClients, grpc.NewClient(member, grpcClientPort))
	}

	la := LeaderAttributes{
		TickClients: tickClients,
	}

	n.LeaderAttr = &la

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

// GetCurLeader returns current leader of the follower.
func (n *Node) GetCurLeader() string {
	return n.FollowerAttr.CurLeader
}

// SetCurLeader sets the current leader of the quorum.
func (n *Node) SetCurLeader(s string) {
	n.FollowerAttr.CurLeader = s
}

// GetAllData returns a copy of all data of the node.
func (n *Node) GetAllData() map[string]string {
	d := n.Data
	return d
}

// GetData returns the value of the key.
func (n *Node) GetData(key string) string {
	return n.Data[key]
}

// SetElectionTimeout refreshes the node's election timeout by re-generate a
// random time duration.
func (n *Node) SetElectionTimeout() {
	n.FollowerAttr.ElectionTimeout = time.Duration(rand.Intn(baseElectionTimeoutMs)+jitterElectionTimeoutMs) * time.Millisecond
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
func (n *Node) AppendEntryC(entry Entry) {
	n.FollowerAttr.EntryC <- entry
}

// AppendTickC appends the tick channel with empty data.
func (n *Node) AppendTickC() {
	n.FollowerAttr.TickC <- struct{}{}
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
			n.SetState(Down)
			close(done)
			return done
		}
	}
}

func (n *Node) startFollower(ctx context.Context) error {
	log.Printf("node %s became Follower", n.ID)
	n.resetFollower(ctx)

	for {
		select {
		case <-time.After(n.FollowerAttr.ElectionTimeout):
			n.SetState(Candidate)
			return nil
		case data := <-n.FollowerAttr.EntryC:
			n.SetElectionTimeout()
			log.Printf("node %s received data %s", n.ID, data)
		case <-n.FollowerAttr.TickC:
			n.SetElectionTimeout()
		}
	}
}

// resetFollower reset follower attributes so that it come up with a fresh
// state.
func (n *Node) resetFollower(ctx context.Context) {
	n.FollowerAttr.EntryC = make(chan Entry)
	n.FollowerAttr.TickC = make(chan struct{})
	n.SetElectionTimeout()
}

func (n *Node) startCandidate(ctx context.Context) error {
	log.Printf("node %s became Candidate", n.ID)

	n.IncrementCurTerm()
	n.SetVotedForTerm(n.GetCurTerm())

	if n.requestVote(ctx) {
		n.SetState(Leader)
	} else {
		n.SetState(Follower)
	}

	return nil
}

func (n *Node) startLeader(ctx context.Context) error {
	log.Printf("node %s became Leader", n.ID)
	n.SetCurLeader(n.GetID())

	for {
		select {
		case <-time.After(tickInterval):
			if accepted := n.appendEntries(ctx, Action_Tick, "", ""); !accepted {
				log.Printf("node %s got rejected by followers", n.ID)
				n.SetState(Follower)
				return nil
			}
		}
	}
}

func (n *Node) requestVote(ctx context.Context) bool {
	votes := 1
	q := len(n.Quorum)

	for _, c := range n.LeaderAttr.TickClients {
		reqCtx, cancel := context.WithTimeout(ctx, grpcReqTimeout)

		accepted, err := c.RequestVote(reqCtx, n.GetID(), n.GetCurTerm())
		if err != nil {
			log.Printf("node %s received error vote response from %s in term %d, consider it as Down", n.GetID(), c.GetAddress(), n.GetCurTerm())
			q--
		} else {
			log.Printf("node %s received votes %v from %s in term %d", n.GetID(), accepted, c.GetAddress(), n.GetCurTerm())

			if accepted {
				votes++
			}
		}

		cancel()
		if votes > len(n.Quorum)/2 {
			return true
		}
	}

	return false
}

func (n *Node) appendEntries(ctx context.Context, action Action, k, v string) bool {
	for _, c := range n.LeaderAttr.TickClients {
		reqCtx, cancel := context.WithTimeout(ctx, grpcReqTimeout)

		curTerm := n.GetCurTerm()

		accepted, term, err := c.AppendEntry(reqCtx, n.GetID(), curTerm, action, k, v)

		if curTerm < term {
			cancel()
			return false
		}

		if err != nil {
			log.Printf("received error response from %s in term %d", c.GetAddress(), n.GetCurTerm())
		}

		cancel()
		if !accepted {
			// TODO: Shouldn't return false, need retry.
			return false
		}
	}
	return true
}

func (n *Node) applyInsert(k, v string) error {
	if _, ok := n.Data[k]; ok {
		return fmt.Errorf("duplicated key %s, use Update instead", k)
	}
	n.Data[k] = v
	return nil
}

func (n *Node) applyGet(k string) (string, error) {
	if v, ok := n.Data[k]; !ok {
		return "", fmt.Errorf("key %s not found", k)
	} else {
		return v, nil
	}
}

func (n *Node) applyUpdate(k, v string) error {
	if _, ok := n.Data[k]; !ok {
		return fmt.Errorf("key %s not found, use Insert instead", k)
	}
	n.Data[k] = v
	return nil
}

func (n *Node) applyDelete(k string) error {
	if _, ok := n.Data[k]; !ok {
		return fmt.Errorf("key %s not found", k)
	}
	delete(n.Data, k)
	return nil
}
