package server

import (
	"github.com/valyala/fasthttp"
)

const (
	ClientPort = 8080

	contentType = "application/json"

	// maxRequestBodySize is the maximum request body size the server reads.
	// Server rejects requests with bodies exceeding this limit.
	maxRequestBodySize = 4 * 1024 * 1024
)

// Handler handles admission review mutation requests.
type Handler struct{}

// Handle processes a mutation admission review request.
//
// The request content type must be application/json.
func (h *Handler) Handle(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetContentType(contentType)
	ctx.SetStatusCode(200)
}

// NewServer creates an HTTP server for serving admission review requests.
func NewServer() (*fasthttp.Server, error) {
	handler := Handler{}

	h := func(ctx *fasthttp.RequestCtx) { handler.Handle(ctx) }

	s := fasthttp.Server{
		Handler:            h,
		MaxRequestBodySize: maxRequestBodySize,
	}
	return &s, nil
}
