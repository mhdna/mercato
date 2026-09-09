package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mhdna/kashi/token"
	"github.com/mhdna/kashi/util"
)

var (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authoizationPayloadKey  = "authorization_payload"

	branchCodeHeaderKey = "x-branch-code"
	branchKeyHeaderKey  = "x-branch-key"
	branchIDKey         = "branch_id"
)

// corsMiddleware reflects back the request's Origin header so the Vite dev
// server (a different origin/port than the API) can call these routes.
func corsMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		origin := ctx.GetHeader("Origin")
		if origin != "" {
			ctx.Header("Access-Control-Allow-Origin", origin)
			ctx.Header("Vary", "Origin")
		}
		ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		ctx.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next()
	}
}

// actorID returns the authenticated user's id for audit stamping. Branch
// (POS) requests carry no user payload, so it returns a null id there.
func actorID(ctx *gin.Context) sql.NullInt64 {
	value, exists := ctx.Get(authoizationPayloadKey)
	if !exists {
		return sql.NullInt64{}
	}
	payload, ok := value.(*token.Payload)
	if !ok || payload.UserID == 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: payload.UserID, Valid: true}
}

func (server *Server) authMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			err := errors.New("authorization header is not provided")
			server.abortError(ctx, http.StatusUnauthorized, err)
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			err := errors.New("invalid authorization header format")
			server.abortError(ctx, http.StatusUnauthorized, err)
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			err := fmt.Errorf("unsupported authorization type: %s", authorizationType)
			server.abortError(ctx, http.StatusUnauthorized, err)
			return
		}

		accessToken := fields[1]
		payload, err := server.tokenMaker.VerifyToken(accessToken)
		if err != nil {
			server.abortError(ctx, http.StatusUnauthorized, err)
			return
		}
		ctx.Set(authoizationPayloadKey, payload)
	}
}

// branchAuthMiddleware authenticates a kashi-pos branch install rather than a
// human user. It's a separate header pair (not Authorization: Bearer, which
// already means "PASETO user token" elsewhere in this codebase) so the two
// schemes can't be confused, and a branch key never expires the way a user
// session does — a branch must keep working after being offline for weeks.
func (server *Server) branchAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		code := ctx.GetHeader(branchCodeHeaderKey)
		key := ctx.GetHeader(branchKeyHeaderKey)
		if code == "" || key == "" {
			server.abortError(ctx, http.StatusUnauthorized, errors.New("branch code and key headers are required"))
			return
		}

		branch, err := server.store.GetBranchByCode(ctx, code)
		if err != nil {
			if err == sql.ErrNoRows {
				err = errors.New("unknown branch")
			}
			server.abortError(ctx, http.StatusUnauthorized, err)
			return
		}

		if !branch.IsActive {
			server.abortError(ctx, http.StatusUnauthorized, errors.New("branch is disabled"))
			return
		}

		if err := util.CheckPassword(key, branch.ApiKeyHash); err != nil {
			server.abortError(ctx, http.StatusUnauthorized, errors.New("invalid branch key"))
			return
		}

		ctx.Set(branchIDKey, branch.ID)

		// Best-effort: a branch shouldn't get a failed request just because
		// this bookkeeping update had a hiccup.
		_ = server.store.UpdateBranchLastSeenAt(ctx, branch.ID)

		ctx.Next()
	}
}
