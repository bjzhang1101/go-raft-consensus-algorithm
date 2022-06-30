// Package node is the package that defines all behaviors of a Raft node.
//
// Behavior of Follower:
//
// Behavior of Candidate:
//
// Behavior of Leader:
// 1.The leader receives clients' requests that contains a command to be
//   executed by the replicated state machines.
// 2.The leader appends the command to its log as a new entry.
// 3.The leader issues AppendEntries RPC calls in parallel to its followers.
// 4.Once the entry is safely replicated, the leader applies the entry to its
//   state machine.
// 5.After the entry is applied, the leader respond to the client.
package node

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/bjzhang1101/raft/client/grpc"
	pb "github.com/bjzhang1101/raft/protobuf"
)

const (
	baseElectionTimeoutMs   = 5000
	jitterElectionTimeoutMs = 5000

	tickInterval = 2500 * time.Millisecond

	grpcReqTimeout = 500 * time.Millisecond

	// applyInterval is the interval between each apply operation.
	applyInterval = 1000 * time.Millisecond
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
	Logs []*pb.Entry

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

// FollowerAttributes is the follower attributes.
//
// TODO: this is most likely useless. Both CurLeader and ElectionTimeout
// can be moved into Node.
type FollowerAttributes struct {
	// CurLeader is the current leader of the quorum.
	CurLeader string

	// ElectionTimeout is the timeout for a follower or a candidate to start
	// (or restart for Candidate) leader election if it receives no
	// communication over the period of time.
	ElectionTimeout time.Duration
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

	// The first index for the meaningful log is 1 instead of 0.
	logs := []*pb.Entry{nil}
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

	n.FollowerAttr = &FollowerAttributes{}

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

// GetLogAction returns the action for the logs[idx].
//
// This function assumes idx is always valid.
func (n *Node) GetLogAction(idx int) pb.Entry_Action {
	return n.Logs[idx].Action
}

// GetLogKey returns the key for the logs[idx].
//
// This function assumes idx is always valid.
func (n *Node) GetLogKey(idx int) string {
	return n.Logs[idx].Key
}

// GetLogValue returns the value for the logs[idx].
//
// This function assumes idx is always valid.
func (n *Node) GetLogValue(idx int) string {
	return n.Logs[idx].Value
}

// AppendLogs appends logs with the given entry.
func (n *Node) AppendLogs(entry *pb.Entry) error {
	if n.State != Leader {
		return fmt.Errorf("append logs operation only allowed for Leader")
	}

	n.Logs = append(n.Logs, entry)
	return nil
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

// SetCommitIdx sets the commit index for the node.
func (n *Node) SetCommitIdx(idx int) {
	n.CommitIdx = idx
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

// Apply applies committed logs.
//
// TODO: this should be able to put in the Start() goroutine.
func (n *Node) Apply(ctx context.Context) {
	for {
		select {
		case <-time.After(applyInterval):
			if n.CommitIdx > n.LastApplied {
				for i := n.LastApplied + 1; i <= n.CommitIdx; i++ {
					action := n.GetLogAction(i)
					key := n.GetLogKey(i)
					value := n.GetLogValue(i)

					if err := n.apply(action, key, value); err != nil {
						log.Printf("failed to apply operation %s for key: %s, value: %s", action, key, value)
						break
					}
					log.Printf("applied operation %s for key: %s, value: %s", action, key, value)
					n.LastApplied++
				}
			}
		}
	}
}

func (n *Node) startFollower(ctx context.Context) error {
	log.Printf("node %s became Follower", n.ID)
	n.SetElectionTimeout()

	for {
		select {
		case <-time.After(n.FollowerAttr.ElectionTimeout):
			n.SetState(Candidate)
			return nil
		}
	}
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
	n.resetLeader(ctx)
	n.SetCurLeader(n.GetID())

	// Send a heartbeat immediately after a node is promoted to be a Leader.
	if accepted, term := n.appendEntries(ctx, pb.Entry_Tick, "", ""); !accepted {
		log.Printf("node %s got rejected by followers", n.ID)
		n.SetState(Follower)
		n.SetCurTerm(term)
		return nil
	}

	for {
		select {
		case <-time.After(tickInterval):
			if accepted, term := n.appendEntries(ctx, pb.Entry_Tick, "", ""); !accepted {
				log.Printf("node %s got rejected by followers", n.ID)
				n.SetState(Follower)
				n.SetCurTerm(term)
				return nil
			}
		}
	}
}

func (n *Node) resetLeader(ctx context.Context) {
	n.LeaderAttr.NextIdx = make(map[string]int)
	n.LeaderAttr.MatchIdx = make(map[string]int)

	for _, member := range n.Quorum {
		// Initialized to leader last log index + 1.
		n.LeaderAttr.NextIdx[member] = len(n.Logs)
		// // Initialized to 0.
		n.LeaderAttr.MatchIdx[member] = 0
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

func (n *Node) sendHeartbeat(ctx context.Context, c *grpc.Client) (bool, int) {
	reqCtx, cancel := context.WithTimeout(ctx, grpcReqTimeout)
	defer cancel()

	curTerm := n.GetCurTerm()

	_, term, err := c.AppendEntries(reqCtx, n.GetID(), curTerm, nil)
	if err != nil {
		log.Printf("received error response from %s in term %d", c.GetAddress(), n.GetCurTerm())
	}

	if curTerm < term {
		return false, term
	}

	return true, n.GetCurTerm()
}

func (n *Node) appendEntries(ctx context.Context, c *grpc.Client) (bool, int) {
	follower := c.GetAddress()

	// TODO: this part need a careful revisit tomorrow.
	if len(n.Logs) < n.LeaderAttr.NextIdx[follower] {
		return n.sendHeartbeat(ctx, c)
	}

	var entries []*pb.Entry
	prevLogIdx := n.LeaderAttr.NextIdx[follower] - 1
	var prevLogTerm int
	if prevLogIdx != 0 {
		n.Logs[prevLogIdx]
	}
	for i := n.LeaderAttr.NextIdx[follower]; i < len(n.Logs); i++ {
		entries = append(entries, &pb.Entry{
			Key:    n.GetLogKey(i),
			Value:  n.GetLogValue(i),
			Action: n.GetLogAction(i),
		})
	}

	reqCtx, cancel := context.WithTimeout(ctx, grpcReqTimeout)
	defer cancel()

	curTerm := n.GetCurTerm()
	accepted, term, err := c.AppendEntries(reqCtx, n.GetID(), curTerm, nil)
	if err != nil {
		log.Printf("received error response from %s in term %d", c.GetAddress(), n.GetCurTerm())
	}

	if curTerm < term {
		return false, term
	}

	// TODO: retry on accepted
	if !accepted {

	}

	return true, n.GetCurTerm()

}

func (n *Node) appendEntriesForAllClient(ctx context.Context) (bool, int) {
	for _, c := range n.LeaderAttr.TickClients {
		reqCtx, cancel := context.WithTimeout(ctx, grpcReqTimeout)

		curTerm := n.GetCurTerm()

		accepted, term, err := c.AppendEntries(reqCtx, n.GetID(), curTerm, action, k, v)

		if curTerm < term {
			cancel()
			return false, term
		}

		if err != nil {
			log.Printf("received error response from %s in term %d", c.GetAddress(), n.GetCurTerm())
		}

		cancel()
		if !accepted {
			// TODO: Shouldn't return false, need retry.
			return false, term
		}
	}
	return true, n.GetCurTerm()
}

func (n *Node) apply(action pb.Entry_Action, k, v string) error {
	switch action {
	case pb.Entry_Insert:
		if _, ok := n.Data[k]; ok {
			return fmt.Errorf("duplicated key %s, use Update instead", k)
		}
		n.Data[k] = v
	case pb.Entry_Update:
		if _, ok := n.Data[k]; !ok {
			return fmt.Errorf("key %s not found, use Insert instead", k)
		}
		n.Data[k] = v
	case pb.Entry_Delete:
		if _, ok := n.Data[k]; !ok {
			return fmt.Errorf("key %s not found", k)
		}
		delete(n.Data, k)
	}
	return nil
}
