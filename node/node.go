// Package node is the package that defines all behaviors of a Raft node.
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

type aeStatus int8

const (
	success aeStatus = iota

	termFailure

	consistencyFailure

	networkFailure
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

	// CurLeader is the current leader of the quorum.
	CurLeader string

	// ElectionTimeout is the timeout for a follower or a candidate to start
	// (or restart for Candidate) leader election if it receives no
	// communication over the period of time.
	ElectionTimeout time.Duration

	// TickC is the channel to make sure the follower received ticks.
	TickC chan struct{}

	// LeaderAttr includes attributes for leader state.
	LeaderAttr *LeaderAttributes
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

	// The first index for the meaningful log is 1 instead of 0. Explicitly
	// set the first placeholder entry's term to be 0 so that the consistency
	// check starts with a clean state.
	logs := []*pb.Entry{{Term: 0}}
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

	rand.Seed(time.Now().UnixNano())
	n.SetElectionTimeout()

	n.TickC = make(chan struct{}, 1)

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
	return n.CurLeader
}

// SetCurLeader sets the current leader of the quorum.
func (n *Node) SetCurLeader(l string) {
	n.CurLeader = l
}

// GetLogs returns the logs of the node.
func (n *Node) GetLogs() []*pb.Entry {
	return n.Logs
}

// GetLog returns the log for the index.
//
// This function assumes idx is always valid.
func (n *Node) GetLog(idx int) *pb.Entry {
	return n.Logs[idx]
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

// GetLogAction returns the action for the logs[idx].
//
// This function assumes idx is always valid.
func (n *Node) GetLogAction(idx int) pb.Entry_Action {
	return n.Logs[idx].Action
}

// SetLogs override the logs by the given new log list.
func (n *Node) SetLogs(logs []*pb.Entry) {
	n.Logs = logs
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

// GetCommitIdx returns the node's commit index.
func (n *Node) GetCommitIdx() int {
	return n.CommitIdx
}

// GetLastApplied returns the node's last applied.
func (n *Node) GetLastApplied() int {
	return n.LastApplied
}

// GetElectionTimeout returns the election timeout for the node.
func (n *Node) GetElectionTimeout() time.Duration {
	return n.ElectionTimeout
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

// SetCommitIdx sets the commit index for the node.
func (n *Node) SetCommitIdx(idx int) {
	n.CommitIdx = idx
}

// AppendTickC appends an empty struct to TickC.
func (n *Node) AppendTickC() {
	n.TickC <- struct{}{}
}

// GetNextIdx returns a copy of NextIdx map.
func (n *Node) GetNextIdx() map[string]int {
	m := make(map[string]int)

	for k, v := range n.LeaderAttr.NextIdx {
		m[k] = v
	}
	return m
}

// GetMatchIdx returns a copy of MatchIdx map.
func (n *Node) GetMatchIdx() map[string]int {
	m := make(map[string]int)

	for k, v := range n.LeaderAttr.MatchIdx {
		m[k] = v
	}
	return m
}

// Start is the main goroutine for a node's main functionality.
//
// TODO: there is a bug here:
// right now if a follower get down and return back, it will fail to receive
// tick and refresh its election timeout.
func (n *Node) Start(ctx context.Context) <-chan struct{} {
	// The goroutine to apply committed log entries.
	go func() {
		for {
			// S1037 - Elaborate way of sleeping
			//
			// Using a select statement with a single case receiving from the
			// result of time.After is a very elaborate way of sleeping that
			// can much simpler be expressed with a simple call to time.Sleep.
			time.Sleep(tickInterval)

			if n.CommitIdx > n.LastApplied {
				for i := n.LastApplied + 1; i <= n.CommitIdx && i < len(n.GetLogs()); i++ {
					action := n.GetLogAction(i)
					key := n.GetLogKey(i)
					value := n.GetLogValue(i)

					if err := n.apply(action, key, value); err != nil {
						log.Printf("failed to apply operation %s for key: %s, value: %s because %v", action, key, value, err)
						// TODO: reconsider error handling when the log
						// entry failed to execute. Right now we just
						// ignore it because we don't want a bad entry to
						// block the entire pipeline.
					} else {
						log.Printf("applied operation %s for key: %s, value: %s", action, key, value)
					}
					n.LastApplied++
				}
			}
		}
	}()

	done := make(chan struct{})
	var err error
	for {
		switch n.State {
		case Follower:
			err = n.startFollower(ctx)
		case Candidate:
			err = n.startCandidate(ctx)
		case Leader:
			n.startLeader(ctx)
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
	n.resetFollower()

	for {
		select {
		case <-time.After(n.GetElectionTimeout()):
			n.SetState(Candidate)
			return nil
		case <-n.TickC:
			n.SetElectionTimeout()
		}
	}
}

func (n *Node) resetFollower() {
	n.SetElectionTimeout()
	n.TickC = make(chan struct{}, 1)
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

func (n *Node) startLeader(ctx context.Context) {
	log.Printf("node %s became Leader", n.ID)
	n.resetLeader(ctx)
	n.SetCurLeader(n.GetID())

	// Send a heartbeat immediately after a node is promoted to be a Leader.
	if !n.sendHeartbeatForAllClient(ctx) {
		log.Printf("node %s is no longer qualified a Leader", n.ID)
		n.SetState(Follower)
		return
	}

	for {
		// S1037 - Elaborate way of sleeping
		//
		// Using a select statement with a single case receiving from the
		// result of time.After is a very elaborate way of sleeping that can
		// much simpler be expressed with a simple call to time.Sleep.
		time.Sleep(tickInterval)

		if !n.appendEntriesForAllClient(ctx) {
			log.Printf("node %s is no longer qualified a Leader", n.ID)
			n.SetState(Follower)
			return
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

func (n *Node) sendHeartbeat(ctx context.Context, c *grpc.Client) aeStatus {
	reqCtx, cancel := context.WithTimeout(ctx, grpcReqTimeout)
	defer cancel()

	curTerm := n.GetCurTerm()

	_, term, err := c.AppendEntries(reqCtx, n.GetID(), int32(curTerm), int32(n.GetCommitIdx()), -1, -1, nil)
	if err != nil {
		log.Printf("received error response from %s in term %d", c.GetAddress(), n.GetCurTerm())
	}

	if curTerm < term {
		n.SetCurTerm(term)
		return termFailure
	}

	return success
}

func (n *Node) appendEntries(ctx context.Context, c *grpc.Client) aeStatus {
	follower := c.GetAddress()

	if len(n.Logs) < n.LeaderAttr.NextIdx[follower] {
		return n.sendHeartbeat(ctx, c)
	}

	var entries []*pb.Entry
	prevLogIdx := n.LeaderAttr.NextIdx[follower] - 1
	prevLogTerm := n.Logs[prevLogIdx].GetTerm()

	for i := n.LeaderAttr.NextIdx[follower]; i < len(n.Logs); i++ {
		entries = append(entries, n.GetLog(i))
	}

	reqCtx, cancel := context.WithTimeout(ctx, grpcReqTimeout)
	defer cancel()

	curTerm := n.GetCurTerm()
	accepted, term, err := c.AppendEntries(reqCtx, n.GetID(), int32(curTerm), int32(n.GetCommitIdx()), int32(prevLogIdx), prevLogTerm, entries)

	// If the follower does not respond, re-create the client and keep
	// retrying forever.
	if err != nil {
		log.Printf("received error response from %s in term %d: %v", c.GetAddress(), n.GetCurTerm(), err)
		return networkFailure
	}

	// If the follower has larger term, return false.
	if curTerm < term {
		n.SetCurTerm(term)
		return termFailure
	}

	// If the consistency check fails, decrement nextIdx and return false.
	if !accepted {
		n.LeaderAttr.NextIdx[follower]--
		return consistencyFailure
	}

	n.LeaderAttr.NextIdx[follower] += len(entries)
	n.LeaderAttr.MatchIdx[follower] = n.LeaderAttr.NextIdx[follower] - 1

	return success

}

func (n *Node) sendHeartbeatForAllClient(ctx context.Context) bool {
	for _, c := range n.LeaderAttr.TickClients {
		switch n.sendHeartbeat(ctx, c) {
		case termFailure:
			return false
		default:
		}
	}
	return true
}

func (n *Node) appendEntriesForAllClient(ctx context.Context) bool {
	consensus := 1

	// TODO: make it in parallel.
	// TODO: add a mutex so that during this operation logs will not change.
	for i, c := range n.LeaderAttr.TickClients {
		switch n.appendEntries(ctx, c) {
		case termFailure:
			return false
		case networkFailure:
			n.LeaderAttr.TickClients[i] = grpc.NewClient(c.GetAddress(), c.GetPort())
		case success:
			consensus++
		default:
			// Do nothing for other failures because retry will handle it.
		}
	}

	if consensus > len(n.Quorum)/2 {
		n.CommitIdx = len(n.Logs) - 1
	}
	return true
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
