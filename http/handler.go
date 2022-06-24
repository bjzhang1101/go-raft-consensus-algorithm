package http

import (
	"encoding/json"

	"github.com/valyala/fasthttp"

	"github.com/bjzhang1101/raft/node"
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
	}{
		Status: h.node.GetState().String(),
		Term:   h.node.GetCurTerm(),
	}

	body, err := json.Marshal(s)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
	}
	ctx.SetBody(body)
}
