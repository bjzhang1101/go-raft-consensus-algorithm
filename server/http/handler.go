package http

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/bjzhang1101/raft/node"
	pb "github.com/bjzhang1101/raft/protobuf"
)

const (
	defaultReqTimeout = 5 * time.Second
)

// Handler handles HTTP requests to the server.
type Handler struct {
	node *node.Node

	requestTimeout time.Duration
}

// NewHandler returns a new HTTP handler.
func NewHandler(node *node.Node) Handler {
	return NewHandlerWithTimeout(node, defaultReqTimeout)
}

// NewHandlerWithTimeout returns a new HTTP handler with request timeout.
func NewHandlerWithTimeout(node *node.Node, reqTimeout time.Duration) Handler {
	return Handler{
		node:           node,
		requestTimeout: reqTimeout,
	}
}

// HandleBlackHole always returns status OK.
func (h *Handler) HandleBlackHole(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
}

// StateResponse is the response body for request /state.
type StateResponse struct {
	State       node.State     `json:"state"`
	Term        int            `json:"term"`
	Leader      string         `json:"leader"`
	CommitIdx   int            `json:"commitIndex"`
	LastApplied int            `json:"lastApplied"`
	Logs        []*pb.Entry    `json:"logs"`
	NextIdx     map[string]int `json:"nextIndex,omitempty"`
	MatchIdx    map[string]int `json:"matchIndex,omitempty"`
}

// HandleState handles the request for path /state to return the node's
// current state.
func (h *Handler) HandleState(ctx *fasthttp.RequestCtx) {
	s := StateResponse{
		State:       h.node.GetState(),
		Term:        h.node.GetCurTerm(),
		Leader:      h.node.GetCurLeader(),
		CommitIdx:   h.node.GetCommitIdx(),
		LastApplied: h.node.GetLastApplied(),
		Logs:        h.node.GetLogs(),
	}

	if h.node.State == node.Leader {
		s.NextIdx = h.node.GetNextIdx()
		s.MatchIdx = h.node.GetMatchIdx()
	}

	body, err := json.Marshal(s)
	if err != nil {
		ctx.Error("internal server error", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.Response.Header.SetContentType(contentType)
	ctx.SetBody(body)
}

// GetAllDataResponse is the response body for request /get_all_data.
type GetAllDataResponse struct {
	Data map[string]string `json:"data"`
}

// HandleGetAllData handles the request for path /get_all_data to return the
// node's current data store.
func (h *Handler) HandleGetAllData(ctx *fasthttp.RequestCtx) {
	s := GetAllDataResponse{Data: h.node.GetAllData()}

	body, err := json.Marshal(s)
	if err != nil {
		ctx.Error("internal server error", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.Response.Header.SetContentType(contentType)
	ctx.SetBody(body)
}

// OperateDataRequest is the request body for request /operate_data.
type OperateDataRequest struct {
	// Key is the key of the log entry.
	Key string `json:"key"`
	// Value is the value of the log entry.
	Value string `json:"value"`
	// Action is the action for the log entry.
	Action string `json:"action"`
}

// OperateDataResponse is the response body for request /operate_data.
type OperateDataResponse struct {
	// Success indicates whether the /operate_data request is successful.
	Success bool `json:"success"`
	// Leader shows the current leader if the request is sent to a follower.
	Leader string `json:"leader"`
}

// HandleOperateData handles the request for path /operate_data to insert,
// get, update and delete data.
func (h *Handler) HandleOperateData(ctx *fasthttp.RequestCtx) {
	if string(ctx.Request.Header.ContentType()) != contentType {
		ctx.Error(fmt.Sprintf("invalid content type: %v", ctx.Request.Header.ContentType()), fasthttp.StatusBadRequest)
		return
	}

	reqBody := ctx.Request.Body()
	var opReq OperateDataRequest

	if err := json.Unmarshal(reqBody, &opReq); err != nil {
		ctx.Error(fmt.Sprintf("failed to parse request body: %v", err), fasthttp.StatusBadRequest)
		return
	}

	action, err := parseAction(opReq.Action)
	if err != nil {
		ctx.Error(fmt.Sprintf("failed to parse action: %v", err), fasthttp.StatusBadRequest)
		return
	}

	key := opReq.Key
	value := opReq.Value
	if len(key) == 0 {
		ctx.Error("invalid empty key", fasthttp.StatusBadRequest)
		return
	}

	if len(value) == 0 && action != pb.Entry_Delete {
		ctx.Error("invalid empty value", fasthttp.StatusBadRequest)
		return
	}

	b := OperateDataResponse{Success: false, Leader: h.node.GetCurLeader()}
	if h.node.GetState() == node.Leader {
		entry := pb.Entry{
			Key:    key,
			Value:  value,
			Action: action,
			Term:   int32(h.node.GetCurTerm()),
		}
		if err := h.node.AppendLogs(&entry); err != nil {
			ctx.Error("internal server error", fasthttp.StatusInternalServerError)
			return
		}

		b.Success = true

		// TODO: This logic is not correct, we can't check data to determine
		// whether the log entry is applied. It should decide based on the
		// operation.
		//
		// Can we check quorum's log entries? If they have the same entry,
		// return true?

		/*
			if v := h.node.GetData(key); value == v {
				b.Success = true
				ctx.SetStatusCode(fasthttp.StatusOK)
				return
			}

			for {
				select {
				case <-time.After(h.requestTimeout):
					ctx.Error("failed to apply entries", fasthttp.StatusInternalServerError)
					return
				case <-time.After(applyStatusInterval):
					if v := h.node.GetData(key); value == v {
						b.Success = true
						ctx.SetStatusCode(fasthttp.StatusOK)
						return
					}
				}
			}

		*/
	}

	body, err := json.Marshal(b)
	if err != nil {
		ctx.Error("interval server failure", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.Response.Header.SetContentType(contentType)
	ctx.SetBody(body)
}

func parseAction(a string) (pb.Entry_Action, error) {
	switch a {
	case "add", "insert":
		return pb.Entry_Insert, nil
	case "update":
		return pb.Entry_Update, nil
	case "del", "delete":
		return pb.Entry_Delete, nil
	default:
		return pb.Entry_InvalidAction, fmt.Errorf("invalid action %s", a)
	}
}
