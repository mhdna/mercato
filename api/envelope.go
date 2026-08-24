package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// envelope wraps every JSON response body in a top-level object keyed by
// name (e.g. {"branch": ...}), rather than returning a bare struct or an
// ad-hoc gin.H per handler. Keeping response shapes predictable this way
// means they can grow later (extra metadata, pagination, etc.) without
// breaking existing clients. Pattern from Alex Edwards' "Let's Go Further".
//
// Scoped to the branch/sync endpoints for now (api/branch.go,
// api/branch_sync.go, and friends) rather than retrofitted across the
// whole api package — that's a separate, larger cleanup.
type envelope map[string]any

func (server *Server) writeJSON(ctx *gin.Context, status int, data envelope) {
	ctx.JSON(status, data)
}

// writeError logs the error server-side whenever it's a 5xx — client errors
// (4xx) are the caller's fault and don't need a server-side trace, but a 500
// means something broke here, and this codebase has no other mechanism that
// records that anywhere today.
func (server *Server) writeError(ctx *gin.Context, status int, err error) {
	if status >= http.StatusInternalServerError {
		log.Printf("error: %v %v: %v", ctx.Request.Method, ctx.Request.URL.Path, err)
	}
	server.writeJSON(ctx, status, envelope{"error": err.Error()})
}

// abortError is writeError for middleware: it also stops the handler chain,
// the same way ctx.AbortWithStatusJSON does.
func (server *Server) abortError(ctx *gin.Context, status int, err error) {
	if status >= http.StatusInternalServerError {
		log.Printf("error: %v %v: %v", ctx.Request.Method, ctx.Request.URL.Path, err)
	}
	ctx.AbortWithStatusJSON(status, envelope{"error": err.Error()})
}
