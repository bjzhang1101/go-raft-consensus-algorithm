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
	applyStatusInterval = 1 * time.Second
	requestTimeout      = 5 * time.Second
)

// Handler handles HTTP requests to the server.
type Handler struct {
	node *node.Node
}

// NewHandler returns a new HTTP handler.
func NewHandler(node *node.Node) Handler {
	return Handler{node: node}
}

// HandleBlackHole always returns status OK.
func (h *Handler) HandleBlackHole(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
}

// HandleState handles the request for path /state to return the node's
// current state.
func (h *Handler) HandleState(ctx *fasthttp.RequestCtx) {
	s := struct {
		Status      string         `json:"state"`
		Term        int            `json:"term"`
		Leader      string         `json:"leader"`
		CommitIdx   int            `json:"commitIndex"`
		LastApplied int            `json:"lastApplied"`
		Logs        []*pb.Entry    `json:"logs"`
		NextIdx     map[string]int `json:"nextIndex,omitempty"`
		MatchIdx    map[string]int `json:"matchIndex,omitempty"`
	}{
		Status:      h.node.GetState().String(),
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

// HandleGetAllData handles the request for path /get_all_data to return the
// node's current data store.
func (h *Handler) HandleGetAllData(ctx *fasthttp.RequestCtx) {
	s := struct {
		Data string `json:"data"`
	}{
		Data: fmt.Sprintf("%v", h.node.GetAllData()),
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

type operateDataRequestBody struct {
	// Key is the key of the log entry.
	Key string `json:"key"`
	// Value is the value of the log entry.
	Value string `json:"value"`
	// Action is the action for the log entry.
	Action string `json:"action"`
}

// HandleOperateData handles the request for path /operate_data to insert,
// get, update and delete data.
func (h *Handler) HandleOperateData(ctx *fasthttp.RequestCtx) {
	if string(ctx.Request.Header.ContentType()) != contentType {
		ctx.Error(fmt.Sprintf("invalid content type: %v", ctx.Request.Header.ContentType()), fasthttp.StatusBadRequest)
		return
	}

	b := struct {
		Success bool   `json:"success"`
		Leader  string `json:"leader"`
	}{}

	reqBody := ctx.Request.Body()
	var opReq operateDataRequestBody

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
		ctx.Error(fmt.Sprint("invalid empty key"), fasthttp.StatusBadRequest)
		return
	}

	if len(value) == 0 && action != pb.Entry_Delete {
		ctx.Error(fmt.Sprint("invalid empty value"), fasthttp.StatusBadRequest)
		return
	}

	if h.node.GetState() != node.Leader {
		b.Success = false
		b.Leader = h.node.GetCurLeader()
	} else {
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

		for {
			select {
			case <-time.After(requestTimeout):
				ctx.Error("failed to apply entries", fasthttp.StatusInternalServerError)
				return
			case <-time.After(applyStatusInterval):
				if v := h.node.GetData(key); value == v {
					b.Success = true
					return
				}
			}
		}
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
