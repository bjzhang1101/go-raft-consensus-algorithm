package node

import pb "github.com/bjzhang1101/raft/protobuf"

type Action = pb.Entry_Action

const (
	Action_Insert = pb.Entry_Insert
	Action_Get    = pb.Entry_Get
	Action_Update = pb.Entry_Update
	Action_Delete = pb.Entry_Delete
	Action_Tick   = pb.Entry_Tick
)

// Entry is the log entry that includes the data and the term.
type Entry struct {
	action Action
	key    string
	value  string
	term   int
}

// NewEntry returns a new entry.
func NewEntry(a Action, k, v string, t int) Entry {
	return Entry{
		action: a,
		key:    k,
		value:  v,
		term:   t,
	}
}
