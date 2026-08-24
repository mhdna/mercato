package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type branchClientRequest struct {
	BranchClientID int64  `json:"branch_client_id" binding:"required"`
	Name           string `json:"name" binding:"required"`
	Phone          string `json:"phone" binding:"required"`
}

// putBranchClient is the up-sync half of client sync: a branch reports a
// client it just created or edited locally. Resolution:
//
//  1. Already linked (branch_id, branch_client_id) -> this branch has
//     pushed this client before, so this is an edit -- update the linked
//     kashi client's fields to match (the branch's edit becomes the new
//     canonical truth, same as it propagating from a live till edit
//     always should).
//  2. Not linked, but an existing kashi client has this exact phone ->
//     the branch thinks this is new but someone else (this branch or
//     another) already created it centrally. Link to the existing
//     client WITHOUT touching its fields -- a phone match alone doesn't
//     mean this branch's name spelling should silently overwrite the
//     established canonical one.
//  3. No link, no phone match anywhere -> genuinely new. Create a kashi
//     client and link it.
//
// Every case ends the same way: broadcast client_updated so other
// branches pick up the (possibly new) canonical record on their next
// catch-up, same as currencies/cashbox_accounts/products today.
func (server *Server) putBranchClient(ctx *gin.Context) {
	branchID := ctx.MustGet(branchIDKey).(int64)

	var req branchClientRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	var client db.Client

	linkedClientID, err := server.store.GetClientLink(ctx, db.GetClientLinkParams{
		BranchID:       branchID,
		BranchClientID: req.BranchClientID,
	})
	switch {
	case err == nil:
		client, err = server.store.UpdateClient(ctx, db.UpdateClientParams{
			ID:    linkedClientID,
			Name:  req.Name,
			Phone: req.Phone,
		})
		if err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
	case err == sql.ErrNoRows:
		existing, phoneErr := server.store.GetClientByPhone(ctx, req.Phone)
		switch {
		case phoneErr == nil:
			client = existing
		case phoneErr == sql.ErrNoRows:
			created, createErr := server.store.CreateClient(ctx, db.CreateClientParams{
				Name:  req.Name,
				Phone: req.Phone,
			})
			if createErr != nil {
				server.writeError(ctx, http.StatusInternalServerError, createErr)
				return
			}
			client = created
		default:
			server.writeError(ctx, http.StatusInternalServerError, phoneErr)
			return
		}
		if err := server.store.UpsertClientLink(ctx, db.UpsertClientLinkParams{
			BranchID:       branchID,
			BranchClientID: req.BranchClientID,
			ClientID:       client.ID,
		}); err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
	default:
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.branchHub.broadcastAll(branchWSMessage{Type: "client_updated"})

	server.writeJSON(ctx, http.StatusOK, envelope{"client": client})
}
