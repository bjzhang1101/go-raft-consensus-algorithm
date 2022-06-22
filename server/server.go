package server

import (
	"github.com/valyala/fasthttp"

	"github.com/bjzhang1101/raft/node"
)

const (
	ClientPort = 8080

	contentType = "application/json"

	// maxRequestBodySize is the maximum request body size the server reads.
	// Server rejects requests with bodies exceeding this limit.
	maxRequestBodySize = 4 * 1024 * 1024

	statusPath = "/status"
)

// NewServer creates an HTTP server serving requests.
func NewServer(node *node.Node) (*fasthttp.Server, error) {
	handler := NewHandler(node)

	h := func(ctx *fasthttp.RequestCtx) {
		switch p := string(ctx.Path()); p {
		case statusPath:
			handler.HandleStatus(ctx)
		default:
			handler.HandleBlackHole(ctx)
		}
	}

	s := fasthttp.Server{
		Handler:            h,
		MaxRequestBodySize: maxRequestBodySize,
	}
	return &s, nil
}
