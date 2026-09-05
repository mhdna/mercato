package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
	"github.com/mhdna/kashi/token"
)

// Branch POS users are an admin-managed, branch-scoped roster synced down to
// kashi-pos read-only (like salespersons). The PIN never lives here -- it's
// pushed via a set_branch_user_pin branch command that the till hashes
// locally. See db/migrations/000073_branch_users.up.sql.

// enqueueSetBranchUserPin queues the plaintext-PIN command for a branch and
// nudges it if connected. Shared by create (initial PIN) and the dedicated
// reset endpoint.
func (server *Server) enqueueSetBranchUserPin(ctx *gin.Context, branchID int64, username, pin string) error {
	payload, err := json.Marshal(map[string]string{"username": username, "pin": pin})
	if err != nil {
		return err
	}

	tokenPayload := ctx.MustGet(authoizationPayloadKey).(*token.Payload)
	user, err := server.store.GetUserByUsername(ctx, tokenPayload.Username)
	if err != nil {
		return err
	}

	command, err := server.store.CreateBranchCommand(ctx, db.CreateBranchCommandParams{
		BranchID: branchID,
		Type:     "set_branch_user_pin",
		Payload:  payload,
		IssuedBy: user.ID,
	})
	if err != nil {
		return err
	}
	server.branchHub.notify(branchID, branchWSMessage{Type: "command", CommandID: command.ID})
	return nil
}

func (server *Server) listBranchUsers(ctx *gin.Context) {
	var uri branchIDRequest
	if err := ctx.ShouldBindUri(&uri); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	users, err := server.store.ListBranchUsersForBranch(ctx, uri.ID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	server.writeJSON(ctx, http.StatusOK, envelope{"branch_users": users})
}

type createBranchUserRequest struct {
	Username     string `json:"username" binding:"required"`
	Role         string `json:"role" binding:"required,oneof=admin cashier"`
	AkuvoxUserID string `json:"akuvox_user_id"`
	IsActive     *bool  `json:"is_active"`
	Pin          string `json:"pin" binding:"required,min=4,max=6,numeric"`
}

func (server *Server) createBranchUser(ctx *gin.Context) {
	var uri branchIDRequest
	if err := ctx.ShouldBindUri(&uri); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	var req createBranchUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	if _, err := server.store.GetBranch(ctx, uri.ID); err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	user, err := server.store.CreateBranchUser(ctx, db.CreateBranchUserParams{
		BranchID:     uri.ID,
		Username:     strings.TrimSpace(req.Username),
		Role:         req.Role,
		AkuvoxUserID: req.AkuvoxUserID,
		IsActive:     isActive,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	if err := server.enqueueSetBranchUserPin(ctx, uri.ID, user.Username, req.Pin); err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	server.branchHub.notify(uri.ID, branchWSMessage{Type: "branch_user_updated"})
	server.writeJSON(ctx, http.StatusOK, envelope{"branch_user": user})
}

type branchUserIDRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type updateBranchUserRequest struct {
	Username     string `json:"username" binding:"required"`
	Role         string `json:"role" binding:"required,oneof=admin cashier"`
	AkuvoxUserID string `json:"akuvox_user_id"`
	IsActive     bool   `json:"is_active"`
}

func (server *Server) updateBranchUser(ctx *gin.Context) {
	var uri branchUserIDRequest
	if err := ctx.ShouldBindUri(&uri); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	var req updateBranchUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	existing, err := server.store.GetBranchUser(ctx, uri.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	user, err := server.store.UpdateBranchUser(ctx, db.UpdateBranchUserParams{
		ID:           uri.ID,
		Username:     strings.TrimSpace(req.Username),
		Role:         req.Role,
		AkuvoxUserID: req.AkuvoxUserID,
		IsActive:     req.IsActive,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.branchHub.notify(existing.BranchID, branchWSMessage{Type: "branch_user_updated"})
	server.writeJSON(ctx, http.StatusOK, envelope{"branch_user": user})
}

func (server *Server) deleteBranchUser(ctx *gin.Context) {
	var uri branchUserIDRequest
	if err := ctx.ShouldBindUri(&uri); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	existing, err := server.store.GetBranchUser(ctx, uri.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	if err := server.store.DeleteBranchUser(ctx, uri.ID); err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.branchHub.notify(existing.BranchID, branchWSMessage{Type: "branch_user_updated"})
	ctx.JSON(http.StatusOK, envelope{"message": "branch user deleted"})
}

type setBranchUserPinRequest struct {
	Pin string `json:"pin" binding:"required,min=4,max=6,numeric"`
}

// setBranchUserPin only enqueues the plaintext-PIN command -- the roster row
// is unchanged (it holds no PIN).
func (server *Server) setBranchUserPin(ctx *gin.Context) {
	var uri branchUserIDRequest
	if err := ctx.ShouldBindUri(&uri); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	var req setBranchUserPinRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	user, err := server.store.GetBranchUser(ctx, uri.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	if err := server.enqueueSetBranchUserPin(ctx, user.BranchID, user.Username, req.Pin); err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, envelope{"message": "pin change queued for the branch"})
}
