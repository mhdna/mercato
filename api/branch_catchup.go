package api

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

var errUnknownSyncEntity = errors.New("unknown entity")

// branchSyncChangesRequest: `since` is an RFC3339 timestamp, or omitted
// entirely to mean "everything" — a branch with an empty local table (a
// fresh install, or one just switching sync on) gets a full snapshot for
// free, since that's just the "since the beginning of time" case of the
// same query. No separate bootstrap mechanism is needed.
type branchSyncChangesRequest struct {
	Entity string `form:"entity" binding:"required"`
	Since  string `form:"since"`
}

// branchSyncChanges is a single generalized catch-up endpoint rather than
// one per entity — it's used both on reconnect and on a plain polling
// interval, and by design returns the same shape a WebSocket delta
// notification would tell the branch to go fetch, so the two mechanisms
// (poll and push) can never disagree about what "current" looks like.
//
// Only two entities exist behind this today (currencies, cashbox_accounts).
// A real per-entity registry is deliberately not built yet — with two
// cases a switch is clearer than a plugin mechanism; extract one once a
// third entity shows what's actually shared.
func (server *Server) branchSyncChanges(ctx *gin.Context) {
	var req branchSyncChangesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	since := time.Time{}
	if req.Since != "" {
		parsed, err := time.Parse(time.RFC3339, req.Since)
		if err != nil {
			server.writeError(ctx, http.StatusBadRequest, err)
			return
		}
		since = parsed
	}

	// Captured before querying, not after: if a write lands between the
	// query running and this line, the branch's next catch-up (using this
	// as its new cursor) will still see it, rather than the client's own
	// clock potentially skipping past a change that landed mid-request.
	serverTime := time.Now().UTC()

	var items any
	var err error
	switch req.Entity {
	case "currencies":
		items, err = server.store.ListCurrenciesUpdatedSince(ctx, since)
	case "cashbox_accounts":
		items, err = server.store.ListCashboxAccountsUpdatedSince(ctx, since)
	case "products":
		items, err = server.store.ListProductVariantsForSync(ctx, since)
	case "clients":
		items, err = server.store.ListClientsUpdatedSince(ctx, since)
	case "branch_targets":
		// Unlike the entities above, targets are branch-specific -- scoped
		// by the authenticated branch's own id rather than global.
		branchID := ctx.MustGet(branchIDKey).(int64)
		items, err = server.store.ListBranchTargetsUpdatedSince(ctx, db.ListBranchTargetsUpdatedSinceParams{
			BranchID:  branchID,
			UpdatedAt: since,
		})
	case "salespersons":
		// Branch-specific like targets -- each branch has its own roster.
		branchID := ctx.MustGet(branchIDKey).(int64)
		items, err = server.store.ListSalespersonsUpdatedSince(ctx, db.ListSalespersonsUpdatedSinceParams{
			BranchID:  sql.NullInt64{Int64: branchID, Valid: true},
			UpdatedAt: since,
		})
	case "branch_users":
		// Branch-specific like salespersons -- each branch's own POS roster.
		// The PIN is never in this feed; it arrives via a set_branch_user_pin
		// command instead.
		branchID := ctx.MustGet(branchIDKey).(int64)
		items, err = server.store.ListBranchUsersUpdatedSince(ctx, db.ListBranchUsersUpdatedSinceParams{
			BranchID:  branchID,
			UpdatedAt: since,
		})
	case "expense_categories":
		// Global like currencies; the query itself restricts to
		// scope='branch' rows -- central-only categories never leave kashi.
		items, err = server.store.ListExpenseCategoriesUpdatedSince(ctx, since)
	case "loan_categories":
		items, err = server.store.ListLoanCategoriesUpdatedSince(ctx, since)
	default:
		server.writeError(ctx, http.StatusBadRequest, errUnknownSyncEntity)
		return
	}
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{
		"entity":      req.Entity,
		"items":       items,
		"server_time": serverTime,
	})
}
