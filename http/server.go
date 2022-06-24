package http

import (
	"github.com/valyala/fasthttp"

	"github.com/bjzhang1101/raft/node"
)

const (
	DefaultPort = 8080

	contentType = "application/json"

	// maxRequestBodySize is the maximum request body size the server reads.
	// Server rejects requests with bodies exceeding this limit.
	maxRequestBodySize = 4 * 1024 * 1024

	statusPath = "/state"

	grpcPath = "/grpc"
)

// NewServer creates an HTTP server serving requests.
func NewServer(node *node.Node, ip string) *fasthttp.Server {
	handler := NewHandler(node)

	h := func(ctx *fasthttp.RequestCtx) {
		switch p := string(ctx.Path()); p {
		case statusPath:
			handler.HandleState(ctx)
		case grpcPath:
			handler.HandlerGRPC(ctx, ip)
		default:
			handler.HandleBlackHole(ctx)
		}
	}

	s := fasthttp.Server{
		Handler:            h,
		MaxRequestBodySize: maxRequestBodySize,
	}
	return &s
}
