package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
	"github.com/mhdna/kashi/token"
	"github.com/sqlc-dev/pqtype"
)

var (
	errUnknownCommandType    = errors.New("unknown or unsupported command type")
	errCommandBranchMismatch = errors.New("command does not belong to this branch")
)

// allowedBranchCommandTypes gates what an admin can enqueue -- adding a new
// remote capability means adding it here AND teaching kashi-pos's command
// executor (see commands.go there) how to run it via a real local function,
// never adding a type that has no corresponding, validated execution path.
var allowedBranchCommandTypes = map[string]bool{
	"update_settings": true,
}

type createBranchCommandRequest struct {
	Type    string          `json:"type" binding:"required"`
	Payload json.RawMessage `json:"payload" binding:"required"`
}

// createBranchCommand is how a human admin asks a branch to do something --
// the enqueue side of the remote command channel. It never talks to
// kashi-pos directly: it writes a pending row (the audit trail) and, if the
// branch happens to be connected right now, nudges it over the WebSocket to
// fetch it immediately. A disconnected or slow-to-notice branch still picks
// it up on its next sync tick (see branchPendingCommands), so this endpoint
// never needs to know or care whether the push actually landed.
func (server *Server) createBranchCommand(ctx *gin.Context) {
	var idReq branchIDRequest
	if err := ctx.ShouldBindUri(&idReq); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	var req createBranchCommandRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	if !allowedBranchCommandTypes[req.Type] {
		server.writeError(ctx, http.StatusBadRequest, errUnknownCommandType)
		return
	}

	if _, err := server.store.GetBranch(ctx, idReq.ID); err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	// The PASETO payload only carries the operator's username (see
	// token.Payload) -- resolve it to a real user id so the audit trail
	// (branch_commands.issued_by) is a durable FK, not a name that could
	// later be renamed or reused.
	tokenPayload := ctx.MustGet(authoizationPayloadKey).(*token.Payload)
	user, err := server.store.GetUserByUsername(ctx, tokenPayload.Username)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	command, err := server.store.CreateBranchCommand(ctx, db.CreateBranchCommandParams{
		BranchID: idReq.ID,
		Type:     req.Type,
		Payload:  req.Payload,
		IssuedBy: user.ID,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.branchHub.notify(idReq.ID, branchWSMessage{Type: "command", CommandID: command.ID})

	server.writeJSON(ctx, http.StatusOK, envelope{"command": command})
}

// listBranchCommands is the admin-facing audit trail for one branch --
// every command ever issued, its payload, and its final result (or that
// it's still pending).
func (server *Server) listBranchCommands(ctx *gin.Context) {
	var idReq branchIDRequest
	if err := ctx.ShouldBindUri(&idReq); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	var req struct {
		PageSize int32 `form:"page_size,default=10" binding:"min=5,max=100"`
		PageID   int32 `form:"page_id,default=0" binding:"min=0"`
	}
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	commands, err := server.store.ListBranchCommands(ctx, db.ListBranchCommandsParams{
		BranchID: idReq.ID,
		Limit:    req.PageSize,
		Offset:   req.PageID,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"branch_commands": commands})
}

// branchPendingCommands is the branch-facing poll side -- called every
// sync tick regardless of whether a WebSocket push was received, the same
// dual-path guarantee every other pushed entity gets (see
// branch_catchup.go). A branch that was offline when a command was issued
// simply finds it here the next time it's back.
func (server *Server) branchPendingCommands(ctx *gin.Context) {
	branchID := ctx.MustGet(branchIDKey).(int64)

	commands, err := server.store.ListPendingBranchCommands(ctx, branchID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"commands": commands})
}

type ackBranchCommandRequest struct {
	Status string          `json:"status" binding:"required,oneof=success failed"`
	Result json.RawMessage `json:"result"`
	Error  string          `json:"error"`
}

// ackBranchCommand is how a branch reports what happened after it executed
// a command -- kashi never re-derives or double-checks the outcome, it
// records exactly what kashi-pos's real service function returned, the
// same "branch is decisive" principle already applied to invoices.
func (server *Server) ackBranchCommand(ctx *gin.Context) {
	var req branchInvoiceIDRequest // {id} — reused, same uri shape
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	var body ackBranchCommandRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	branchID := ctx.MustGet(branchIDKey).(int64)
	command, err := server.store.GetBranchCommand(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	if command.BranchID != branchID {
		server.writeError(ctx, http.StatusForbidden, errCommandBranchMismatch)
		return
	}

	var errText sql.NullString
	if body.Error != "" {
		errText = sql.NullString{String: body.Error, Valid: true}
	}
	var result pqtype.NullRawMessage
	if len(body.Result) > 0 {
		result = pqtype.NullRawMessage{RawMessage: body.Result, Valid: true}
	}

	updated, err := server.store.CompleteBranchCommand(ctx, db.CompleteBranchCommandParams{
		ID:     req.ID,
		Status: body.Status,
		Result: result,
		Error:  errText,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"command": updated})
}
