package http

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/valyala/fasthttp"

	pb "github.com/bjzhang1101/raft/protobuf"

	"github.com/bjzhang1101/raft/node"
)

const (
	applyStatusInterval = 1 * time.Second
	requestTimeout      = 5 * time.Second
)

var (
	counter = 1
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
	ctx.Response.Header.SetContentType(contentType)
	ctx.SetStatusCode(fasthttp.StatusOK)

	s := struct {
		Status string `json:"status"`
		Term   int    `json:"term"`
		Leader string `json:"leader"`
	}{
		Status: h.node.GetState().String(),
		Term:   h.node.GetCurTerm(),
		Leader: h.node.GetCurLeader(),
	}

	body, err := json.Marshal(s)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
	}
	ctx.SetBody(body)
}

// HandleGetAllData handles the request for path /get_all_data to return the
// node's current data store.
func (h *Handler) HandleGetAllData(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetContentType(contentType)
	ctx.SetStatusCode(fasthttp.StatusOK)

	m := h.node.GetAllData()

	dataBuilder := strings.Builder{}

	for k, v := range m {
		dataBuilder.WriteString(fmt.Sprintf("%s: %s\n", k, v))
	}

	s := struct {
		Data string `json:"data"`
	}{
		Data: dataBuilder.String(),
	}

	body, err := json.Marshal(s)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
	}
	ctx.SetBody(body)
}

// HandleOperateData handles the request for path /operate_data to insert,
// get, update and delete data.
func (h *Handler) HandleOperateData(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetContentType(contentType)
	ctx.SetStatusCode(fasthttp.StatusOK)

	key := fmt.Sprintf("key-%d", counter)
	value := fmt.Sprintf("value-%d", counter)

	b := struct {
		Success bool   `json:"success"`
		Leader  string `json:"leader"`
	}{}

	if h.node.GetState() != node.Leader {
		b.Success = false
		b.Leader = h.node.GetCurLeader()
	} else {
		entry := pb.Entry{
			Key:    key,
			Value:  value,
			Action: pb.Entry_Insert,
			Term:   int32(h.node.GetCurTerm()),
		}
		if err := h.node.AppendLogs(&entry); err != nil {
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			return
		}

		b.Success = true

		for {
			select {
			case <-time.After(requestTimeout):
				ctx.SetStatusCode(fasthttp.StatusInternalServerError)

				entry := pb.Entry{
					Key:    key,
					Value:  value,
					Action: pb.Entry_Delete,
					Term:   int32(h.node.GetCurTerm()),
				}
				// Is this necessary to add an entry to delete this data if timeout?
				h.node.AppendLogs(&entry)
				return
			case <-time.After(applyStatusInterval):
				if v := h.node.GetData(key); value == v {
					counter++
					return
				}
			}
		}
	}

	body, err := json.Marshal(b)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetBody(body)
}
