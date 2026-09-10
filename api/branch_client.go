package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type branchClientRequest struct {
	BranchClientID int64  `json:"branch_client_id" binding:"required"`
	Name           string `json:"name" binding:"required"`
	// Not "required": a legacy till may hold a client with a blank/garbage
	// phone. Rather than 400 (retried forever) or create an unmatchable
	// central row, that case is staged as an 'invalid_phone' conflict.
	Phone string `json:"phone"`
}

// validBranchPhone reports whether a branch-reported phone is usable as the
// global clients.phone key: at least 6 digits after stripping the usual
// formatting characters.
func validBranchPhone(phone string) bool {
	digits := 0
	for _, r := range phone {
		if r >= '0' && r <= '9' {
			digits++
		}
	}
	return digits >= 6
}

// putBranchClient is the up-sync half of client sync: a branch reports a
// client it just created or edited locally. Resolution:
//
//  1. Already linked (branch_id, branch_client_id) -> an edit. Update the
//     linked kashi client to match; client_type is never touched.
//  2. Not linked, phone matches an existing kashi client -> link to it so
//     this branch's invoices attribute immediately. If the names differ,
//     stage a 'phone_name_mismatch' conflict for an admin rather than
//     silently keeping either spelling.
//  3. Not linked, phone unusable -> stage an 'invalid_phone' conflict;
//     create and link nothing. Still a 200 so the outbox row is consumed.
//  4. Not linked, no phone match -> genuinely new. Create and link.
func (server *Server) putBranchClient(ctx *gin.Context) {
	branchID := ctx.MustGet(branchIDKey).(int64)

	var req branchClientRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	req.Phone = strings.TrimSpace(req.Phone)
	payloadJSON, _ := json.Marshal(req)
	refStr := strconv.FormatInt(req.BranchClientID, 10)

	linkedClientID, err := server.store.GetClientLink(ctx, db.GetClientLinkParams{
		BranchID:       branchID,
		BranchClientID: req.BranchClientID,
	})
	switch {
	case err == nil:
		// An edit of an already-linked client. client_type is never touched
		// -- an admin may have reclassified this client to wholesale, and a
		// branch till has no business reverting that.
		linked, getErr := server.store.GetClient(ctx, linkedClientID)
		if getErr != nil {
			server.writeError(ctx, http.StatusInternalServerError, getErr)
			return
		}
		client, updErr := server.store.UpdateClient(ctx, db.UpdateClientParams{
			ID:         linkedClientID,
			Name:       req.Name,
			Phone:      req.Phone,
			ClientType: linked.ClientType,
		})
		if updErr != nil {
			server.writeError(ctx, http.StatusInternalServerError, updErr)
			return
		}
		server.branchHub.broadcastAll(branchWSMessage{Type: "client_updated"})
		server.writeJSON(ctx, http.StatusOK, envelope{"client": client})
		return

	case err == sql.ErrNoRows:
		// Not linked yet.
		if !validBranchPhone(req.Phone) {
			server.stageBranchSyncConflict(ctx, branchID, "client", refStr, "invalid_phone", payloadJSON, sql.NullInt64{})
			server.writeJSON(ctx, http.StatusOK, envelope{"conflict": true, "conflict_kind": "invalid_phone"})
			return
		}

		existing, phoneErr := server.store.GetClientByPhone(ctx, req.Phone)
		switch {
		case phoneErr == nil:
			if linkErr := server.store.UpsertClientLink(ctx, db.UpsertClientLinkParams{
				BranchID:       branchID,
				BranchClientID: req.BranchClientID,
				ClientID:       existing.ID,
			}); linkErr != nil {
				server.writeError(ctx, http.StatusInternalServerError, linkErr)
				return
			}
			if strings.TrimSpace(existing.Name) != strings.TrimSpace(req.Name) {
				server.stageBranchSyncConflict(ctx, branchID, "client", refStr, "phone_name_mismatch", payloadJSON,
					sql.NullInt64{Int64: existing.ID, Valid: true})
				server.branchHub.broadcastAll(branchWSMessage{Type: "client_updated"})
				server.writeJSON(ctx, http.StatusOK, envelope{
					"client": existing, "conflict": true, "conflict_kind": "phone_name_mismatch",
				})
				return
			}
			server.branchHub.broadcastAll(branchWSMessage{Type: "client_updated"})
			server.writeJSON(ctx, http.StatusOK, envelope{"client": existing})
			return

		case phoneErr == sql.ErrNoRows:
			created, createErr := server.store.CreateClient(ctx, db.CreateClientParams{
				Name:       req.Name,
				Phone:      req.Phone,
				ClientType: "retail",
			})
			var client db.Client
			switch {
			case createErr == nil:
				client = created
			case isUniqueViolation(createErr):
				// A concurrent push won the race on the global phone index.
				// Resolve to the winner rather than 500ing the loser forever.
				won, lookupErr := server.store.GetClientByPhone(ctx, req.Phone)
				if lookupErr != nil {
					server.writeError(ctx, http.StatusInternalServerError, lookupErr)
					return
				}
				client = won
			default:
				server.writeError(ctx, http.StatusInternalServerError, createErr)
				return
			}
			if linkErr := server.store.UpsertClientLink(ctx, db.UpsertClientLinkParams{
				BranchID:       branchID,
				BranchClientID: req.BranchClientID,
				ClientID:       client.ID,
			}); linkErr != nil {
				server.writeError(ctx, http.StatusInternalServerError, linkErr)
				return
			}
			server.branchHub.broadcastAll(branchWSMessage{Type: "client_updated"})
			server.writeJSON(ctx, http.StatusOK, envelope{"client": client})
			return

		default:
			server.writeError(ctx, http.StatusInternalServerError, phoneErr)
			return
		}

	default:
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
}
