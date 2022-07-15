package http

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/valyala/fasthttp"

	"github.com/bjzhang1101/raft/node"
	pb "github.com/bjzhang1101/raft/protobuf"
)

func TestHandleBlackHole(t *testing.T) {
	cases := []struct {
		status int
	}{
		{
			status: 200,
		},
	}

	for i, c := range cases {
		var ctx fasthttp.RequestCtx

		h := NewHandler(nil)
		h.HandleBlackHole(&ctx)
		if c.status != ctx.Response.StatusCode() {
			t.Errorf("[%d] wanted http status %d, but got %d", i, c.status, ctx.Response.StatusCode())
		}
	}
}

func TestHandleState(t *testing.T) {
	cases := []struct {
		state       node.State
		term        int
		leader      string
		commitIdx   int
		lastApplied int
		logs        []*pb.Entry
		nextIdx     map[string]int
		matchIdx    map[string]int
		status      int
		out         StateResponse
	}{
		{
			state:       node.Follower,
			term:        1,
			leader:      "raft-01.raft.svc",
			commitIdx:   1,
			lastApplied: 1,
			logs: []*pb.Entry{{
				Key:    "key-1",
				Value:  "value-1",
				Action: pb.Entry_Insert,
				Term:   1,
			}},
			status: 200,
			out: StateResponse{
				State:       node.Follower,
				Term:        1,
				Leader:      "raft-01.raft.svc",
				CommitIdx:   1,
				LastApplied: 1,
				Logs: []*pb.Entry{{
					Key:    "key-1",
					Value:  "value-1",
					Action: pb.Entry_Insert,
					Term:   1,
				}},
			},
		},
		{
			state:       node.Leader,
			term:        1,
			leader:      "raft-01.raft.svc",
			commitIdx:   1,
			lastApplied: 1,
			logs: []*pb.Entry{{
				Key:    "key-1",
				Value:  "value-1",
				Action: pb.Entry_Insert,
				Term:   1,
			}},
			nextIdx:  map[string]int{"raft-02.raft.svc": 1, "raft-03.raft.svc": 1},
			matchIdx: map[string]int{"raft-02.raft.svc": 1, "raft-03.raft.svc": 1},
			status:   200,
			out: StateResponse{
				State:       node.Leader,
				Term:        1,
				Leader:      "raft-01.raft.svc",
				CommitIdx:   1,
				LastApplied: 1,
				Logs: []*pb.Entry{{
					Key:    "key-1",
					Value:  "value-1",
					Action: pb.Entry_Insert,
					Term:   1,
				}},
				NextIdx:  map[string]int{"raft-02.raft.svc": 1, "raft-03.raft.svc": 1},
				MatchIdx: map[string]int{"raft-02.raft.svc": 1, "raft-03.raft.svc": 1},
			},
		},
	}

	for i, c := range cases {
		var ctx fasthttp.RequestCtx

		n := node.Node{
			State:       c.state,
			CurTerm:     c.term,
			Logs:        c.logs,
			CommitIdx:   c.commitIdx,
			LastApplied: c.lastApplied,
			CurLeader:   c.leader,
			LeaderAttr:  &node.LeaderAttributes{},
		}

		if len(c.nextIdx) != 0 {
			n.LeaderAttr.NextIdx = c.nextIdx
		}

		if len(c.matchIdx) != 0 {
			n.LeaderAttr.MatchIdx = c.matchIdx
		}

		h := NewHandler(&n)
		h.HandleState(&ctx)
		if c.status != ctx.Response.StatusCode() {
			t.Errorf("[%d] wanted http status %d, but got %d", i, c.status, ctx.Response.StatusCode())
			continue
		}

		if "application/json" != string(ctx.Response.Header.ContentType()) {
			t.Errorf("[%d] wanted content type application/json, but got %s", i, ctx.Response.Header.ContentType())
			continue
		}

		var out StateResponse
		if err := json.Unmarshal(ctx.Response.Body(), &out); err != nil {
			t.Fatalf("failed to unmarshal response body: %v", err)
		}
		if diff := cmp.Diff(c.out, out, cmpopts.IgnoreUnexported(pb.Entry{})); len(diff) > 0 {
			t.Errorf("[%d] result mismatch (-want +got)\n%s", i, diff)
		}
	}
}

func TestHandleGetAllData(t *testing.T) {
	cases := []struct {
		data   map[string]string
		status int
		out    GetAllDataResponse
	}{
		{
			data:   map[string]string{"key-1": "value-1", "key-2": "value-2"},
			status: 200,
			out:    GetAllDataResponse{Data: map[string]string{"key-1": "value-1", "key-2": "value-2"}},
		},
	}

	for i, c := range cases {
		var ctx fasthttp.RequestCtx

		n := node.Node{Data: map[string]string{"key-1": "value-1", "key-2": "value-2"}}
		h := NewHandler(&n)
		h.HandleGetAllData(&ctx)

		if c.status != ctx.Response.StatusCode() {
			t.Errorf("[%d] wanted http status %d, but got %d", i, c.status, ctx.Response.StatusCode())
			continue
		}

		if "application/json" != string(ctx.Response.Header.ContentType()) {
			t.Errorf("[%d] wanted content type application/json, but got %s", i, ctx.Response.Header.ContentType())
			continue
		}

		var out GetAllDataResponse
		if err := json.Unmarshal(ctx.Response.Body(), &out); err != nil {
			t.Fatalf("failed to unmarshal response body: %v", err)
		}
		if diff := cmp.Diff(c.out, out, cmpopts.IgnoreUnexported(pb.Entry{})); len(diff) > 0 {
			t.Errorf("[%d] result mismatch (-want +got)\n%s", i, diff)
		}
	}
}

func TestHandleOperateData(t *testing.T) {
	cases := []struct {
		reqTimeout  time.Duration
		contentType string
		in          OperateDataRequest
		state       node.State
		term        int
		leader      string
		data        map[string]string
		status      int
		out         OperateDataResponse
	}{
		{
			state:       node.Follower,
			contentType: "application/json",
			in: OperateDataRequest{
				Key:    "k",
				Value:  "v",
				Action: "add",
			},
			term:   1,
			leader: "raft-01.raft.svc",
			status: 200,
			out: OperateDataResponse{
				Success: false,
				Leader:  "raft-01.raft.svc",
			},
		},
		{
			state:       node.Leader,
			contentType: "application/json",
			in: OperateDataRequest{
				Key:    "k",
				Value:  "v",
				Action: "add",
			},
			status: 200,
			out: OperateDataResponse{
				Success: true,
				Leader:  "raft-01.raft.svc",
			},
		},
		{
			state:       node.Leader,
			contentType: "application/json",
			in: OperateDataRequest{
				Key:    "k",
				Action: "del",
			},
			status: 200,
			out: OperateDataResponse{
				Success: true,
				Leader:  "raft-01.raft.svc",
			},
		},
		// Invalid content type.
		{
			state:       node.Leader,
			contentType: "text/plain; charset=utf-8",
			status:      400,
		},
		// Invalid request: invalid action.
		{
			state:       node.Leader,
			contentType: "application/json",
			in: OperateDataRequest{
				Key:    "k",
				Value:  "v",
				Action: "fake",
			},
			status: 400,
		},
		// Invalid request: invalid key.
		{
			state:       node.Leader,
			contentType: "application/json",
			in: OperateDataRequest{
				Key:    "",
				Value:  "v",
				Action: "add",
			},
			status: 400,
		},
		// Invalid request: invalid value.
		{
			state:       node.Leader,
			contentType: "application/json",
			in: OperateDataRequest{
				Key:    "k",
				Value:  "",
				Action: "add",
			},
			status: 400,
		},
	}

	for i, c := range cases {
		in, err := json.Marshal(c.in)
		if err != nil {
			t.Fatalf("failed to encode request body: %v", err)
		}

		var ctx fasthttp.RequestCtx
		ctx.Request.Header.SetContentType(c.contentType)
		ctx.Request.SetBody(in)

		n := node.Node{
			State:     c.state,
			CurTerm:   1,
			CurLeader: "raft-01.raft.svc",
			Logs:      []*pb.Entry{},
		}

		h := NewHandlerWithTimeout(&n, c.reqTimeout)
		h.HandleOperateData(&ctx)

		if c.status != ctx.Response.StatusCode() {
			t.Errorf("[%d] wanted http status %d, but got %d", i, c.status, ctx.Response.StatusCode())
			continue
		}

		if c.status != fasthttp.StatusOK {
			continue
		}

		if "application/json" != string(ctx.Response.Header.ContentType()) {
			t.Errorf("[%d] wanted content type application/json, but got %s", i, ctx.Response.Header.ContentType())
			continue
		}

		var out OperateDataResponse
		if err := json.Unmarshal(ctx.Response.Body(), &out); err != nil {
			t.Fatalf("failed to unmarshal response body: %v", err)
		}

		if diff := cmp.Diff(c.out, out, cmpopts.IgnoreUnexported(pb.Entry{})); len(diff) > 0 {
			t.Errorf("[%d] result mismatch (-want +got)\n%s", i, diff)
		}
	}
}
