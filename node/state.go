package node

type State int8

const (
	Follower State = iota

	Candidate

	Leader

	Down
)

// String returns string type of Status.
func (s State) String() string {
	switch s {
	case Follower:
		return "follower"
	case Candidate:
		return "candidate"
	case Leader:
		return "leader"
	case Down:
		return "down"
	default:
		return "invalid status"
	}
}
