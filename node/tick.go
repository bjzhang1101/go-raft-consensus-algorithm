package node

// Tick stores the information of each tick.
//
// Tick can have an empty Log serving as a heartbeat only.
type Tick struct {
	// ID is the ID of the leader.
	ID string

	// Term is the leader or candidate's current term.
	Term int

	// Log is the data that needs to be replicated across servers.
	Log string

	// VoteRequest indicates whether the tick is for leader election.
	VoteRequest bool
}

// TickResp is the response for a TickRequest.
type TickResp struct {
	// Consensus indicates whether the TickRequest is consensed.
	Consensus bool
}

// GetConsensus returns whether the TickRequest is consensed.
func (resp TickResp) GetConsensus() bool {
	return resp.Consensus
}
