package http

import (
	"log"

	"github.com/valyala/fasthttp"

	"github.com/bjzhang1101/raft/node"
)

const (
	DefaultPort = 8080

	contentType = "application/json"

	// maxRequestBodySize is the maximum request body size the server reads.
	// Server rejects requests with bodies exceeding this limit.
	maxRequestBodySize = 4 * 1024 * 1024

	statusPath      = "/state"
	getAllDataPath  = "/get_all_data"
	operateDataPath = "/operate_data"
)

// NewServer creates an HTTP server serving requests.
func NewServer(node *node.Node) *fasthttp.Server {
	log.Printf("starting http server listening on :%d", DefaultPort)
	handler := NewHandler(node)

	h := func(ctx *fasthttp.RequestCtx) {
		switch p := string(ctx.Path()); p {
		case statusPath:
			handler.HandleState(ctx)
		case getAllDataPath:
			handler.HandleGetAllData(ctx)
		case operateDataPath:
			handler.HandleOperateData(ctx)
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
