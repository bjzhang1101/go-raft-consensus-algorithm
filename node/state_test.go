package node_test

import (
	"testing"

	"github.com/bjzhang1101/raft/node"
)

func TestString(t *testing.T) {
	cases := []struct {
		in  node.State
		out string
	}{
		{
			in:  node.Follower,
			out: "follower",
		},
		{
			in:  node.Candidate,
			out: "candidate",
		},
		{
			in:  node.Leader,
			out: "leader",
		},
		{
			in:  node.Down,
			out: "down",
		},
	}

	for i, c := range cases {
		out := c.in.String()

		if c.out != out {
			t.Errorf("[%d] wanted %s, but got %s", i, c.out, out)
		}
	}
}
