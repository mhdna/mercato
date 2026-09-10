package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

var errUnknownSyncEntity = errors.New("unknown entity")

// syncChangesDefaultLimit / syncChangesMaxLimit bound a paged catch-up
// request. A branch that omits `limit` entirely gets the pre-pagination
// behaviour: one unbounded query, no next_cursor -- kept so an older
// kashi-pos build keeps working unchanged.
const (
	syncChangesDefaultLimit = 500
	syncChangesMaxLimit     = 1000
)

// pagedSyncEntities are the only entities whose tables are unbounded in
// principle (the whole retail client list, the whole variant catalog), so
// they're the only ones that honour `limit` + `cursor`. Everything else
// (currencies, cashbox_accounts, categories, and the per-branch feeds) is
// small by construction and still answered in a single shot.
var pagedSyncEntities = map[string]bool{
	"products": true,
	"clients":  true,
}

// branchSyncChangesRequest: `since` is an RFC3339 timestamp, or omitted
// entirely to mean "everything" — a branch with an empty local table (a
// fresh install, or one just switching sync on) gets a full snapshot for
// free, since that's just the "since the beginning of time" case of the
// same query. No separate bootstrap mechanism is needed.
//
// `limit` (paged entities only) caps the rows returned; when the response
// fills that limit it also carries `next_cursor`, an opaque
// "<updated_at>|<id>" the branch feeds back as `cursor` to get the next
// page. `cursor`, when present, supersedes `since`.
type branchSyncChangesRequest struct {
	Entity string `form:"entity" binding:"required"`
	Since  string `form:"since"`
	Limit  int    `form:"limit"`
	Cursor string `form:"cursor"`
}

// syncCursor is the (updated_at, id) keyset a paged catch-up walks. id
// breaks ties between rows sharing an exact updated_at so a page boundary
// landing between two such rows never drops or repeats one.
type syncCursor struct {
	updatedAt time.Time
	id        int64
}

func (c syncCursor) encode() string {
	return c.updatedAt.UTC().Format(time.RFC3339Nano) + "|" + strconv.FormatInt(c.id, 10)
}

func parseSyncCursor(raw string) (syncCursor, error) {
	parts := strings.SplitN(raw, "|", 2)
	if len(parts) != 2 {
		return syncCursor{}, fmt.Errorf("malformed cursor %q", raw)
	}
	ts, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return syncCursor{}, fmt.Errorf("malformed cursor timestamp: %w", err)
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return syncCursor{}, fmt.Errorf("malformed cursor id: %w", err)
	}
	return syncCursor{updatedAt: ts, id: id}, nil
}

// branchSyncChanges is a single generalized catch-up endpoint rather than
// one per entity — it's used both on reconnect and on a plain polling
// interval, and by design returns the same shape a WebSocket delta
// notification would tell the branch to go fetch, so the two mechanisms
// (poll and push) can never disagree about what "current" looks like.
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

	paged := pagedSyncEntities[req.Entity] && req.Limit > 0
	nextCursor := ""

	var items any
	var err error
	switch req.Entity {
	case "currencies":
		items, err = server.store.ListCurrenciesUpdatedSince(ctx, since)
	case "cashbox_accounts":
		items, err = server.store.ListCashboxAccountsUpdatedSince(ctx, since)
	case "products":
		if paged {
			var rows []db.ListProductVariantsForSyncPagedRow
			cur := syncCursor{updatedAt: since}
			cur, err = resolveCursor(req.Cursor, cur)
			if err != nil {
				server.writeError(ctx, http.StatusBadRequest, err)
				return
			}
			limit := clampSyncLimit(req.Limit)
			rows, err = server.store.ListProductVariantsForSyncPaged(ctx, db.ListProductVariantsForSyncPagedParams{
				AfterUpdatedAt: cur.updatedAt,
				AfterID:        cur.id,
				RowLimit:       int32(limit),
			})
			if err == nil && len(rows) == limit {
				last := rows[len(rows)-1]
				nextCursor = syncCursor{updatedAt: last.UpdatedAt, id: last.ID}.encode()
			}
			items = rows
		} else {
			items, err = server.store.ListProductVariantsForSync(ctx, since)
		}
	case "clients":
		if paged {
			var rows []db.Client
			cur := syncCursor{updatedAt: since}
			cur, err = resolveCursor(req.Cursor, cur)
			if err != nil {
				server.writeError(ctx, http.StatusBadRequest, err)
				return
			}
			limit := clampSyncLimit(req.Limit)
			rows, err = server.store.ListClientsUpdatedSincePaged(ctx, db.ListClientsUpdatedSincePagedParams{
				AfterUpdatedAt: cur.updatedAt,
				AfterID:        cur.id,
				RowLimit:       int32(limit),
			})
			if err == nil && len(rows) == limit {
				last := rows[len(rows)-1]
				nextCursor = syncCursor{updatedAt: last.UpdatedAt, id: last.ID}.encode()
			}
			items = rows
		} else {
			items, err = server.store.ListClientsUpdatedSince(ctx, since)
		}
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
		"next_cursor": nextCursor,
	})
}

func clampSyncLimit(limit int) int {
	if limit <= 0 {
		return syncChangesDefaultLimit
	}
	if limit > syncChangesMaxLimit {
		return syncChangesMaxLimit
	}
	return limit
}

// resolveCursor returns the keyset to page from: the explicit `cursor`
// query param when the branch sent one (mid-walk), otherwise the
// since-derived starting keyset (id 0, first page).
func resolveCursor(raw string, sinceDerived syncCursor) (syncCursor, error) {
	if raw == "" {
		return sinceDerived, nil
	}
	return parseSyncCursor(raw)
}

// --- reconcile: "which of these do you already have?" -----------------------

const branchReconcileMaxRefs = 500

var errTooManyReconcileRefs = fmt.Errorf("too many refs in one request (max %d)", branchReconcileMaxRefs)

type branchReconcileRequest struct {
	Entity string   `json:"entity" binding:"required"`
	Refs   []string `json:"refs"`
}

// branchReconcile answers, for a batch of refs the branch believes it has
// synced, which ones kashi actually holds. The branch queues (or re-queues)
// only the difference. This is what powers "sync everything now" and, run
// after the central URL is re-pointed, what lets the branch tell whether
// the new server is already up to date -- without blindly re-pushing
// everything and leaning on (branch_id, client_ref) idempotency to absorb
// the flood.
//
// Everything is scoped to the authenticated branch (never a body field).
// Products are the exception: a barcode is globally unique in kashi, so
// "does this barcode resolve to a central variant" is a global question.
func (server *Server) branchReconcile(ctx *gin.Context) {
	branchID := ctx.MustGet(branchIDKey).(int64)

	var req branchReconcileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	if len(req.Refs) > branchReconcileMaxRefs {
		server.writeError(ctx, http.StatusBadRequest, errTooManyReconcileRefs)
		return
	}
	if len(req.Refs) == 0 {
		server.writeJSON(ctx, http.StatusOK, envelope{"missing": []string{}})
		return
	}

	present := make(map[string]bool, len(req.Refs))
	var err error

	switch req.Entity {
	case "invoices":
		var found []string
		found, err = server.store.BranchInvoiceRefsPresent(ctx, db.BranchInvoiceRefsPresentParams{
			BranchID: branchID,
			Column2:  req.Refs,
		})
		for _, r := range found {
			present[r] = true
		}
	case "expenses":
		var found []string
		found, err = server.store.BranchExpenseRefsPresent(ctx, db.BranchExpenseRefsPresentParams{
			BranchID: branchID,
			Column2:  req.Refs,
		})
		for _, r := range found {
			present[r] = true
		}
	case "loans":
		var found []sql.NullString
		found, err = server.store.BranchLoanRefsPresent(ctx, db.BranchLoanRefsPresentParams{
			BranchID: sql.NullInt64{Int64: branchID, Valid: true},
			Column2:  req.Refs,
		})
		for _, r := range found {
			if r.Valid {
				present[r.String] = true
			}
		}
	case "clients":
		// Refs are branch_client_id values as text; client_links keys on
		// (branch_id, branch_client_id).
		ids := make([]int64, 0, len(req.Refs))
		for _, r := range req.Refs {
			n, convErr := strconv.ParseInt(r, 10, 64)
			if convErr != nil {
				server.writeError(ctx, http.StatusBadRequest, fmt.Errorf("client ref %q is not an integer", r))
				return
			}
			ids = append(ids, n)
		}
		var found []int64
		found, err = server.store.BranchClientLinksPresent(ctx, db.BranchClientLinksPresentParams{
			BranchID: branchID,
			Column2:  ids,
		})
		for _, id := range found {
			present[strconv.FormatInt(id, 10)] = true
		}
	case "products":
		// Refs are barcodes; barcode is globally unique, no branch scope.
		var found []string
		found, err = server.store.ProductBarcodesPresent(ctx, req.Refs)
		for _, r := range found {
			present[r] = true
		}
	case "shift_closes":
		var found []string
		found, err = server.store.BranchShiftRefsPresent(ctx, db.BranchShiftRefsPresentParams{
			BranchID: branchID, Column2: req.Refs,
		})
		for _, r := range found {
			present[r] = true
		}
	case "visitor_events":
		var found []string
		found, err = server.store.BranchVisitorEventRefsPresent(ctx, db.BranchVisitorEventRefsPresentParams{
			BranchID: branchID, Column2: req.Refs,
		})
		for _, r := range found {
			present[r] = true
		}
	case "attendance_events":
		var found []string
		found, err = server.store.BranchAttendanceEventRefsPresent(ctx, db.BranchAttendanceEventRefsPresentParams{
			BranchID: branchID, Column2: req.Refs,
		})
		for _, r := range found {
			present[r] = true
		}
	case "attendance_changes":
		var found []string
		found, err = server.store.BranchAttendanceChangeRefsPresent(ctx, db.BranchAttendanceChangeRefsPresentParams{
			BranchID: branchID, Column2: req.Refs,
		})
		for _, r := range found {
			present[r] = true
		}
	default:
		server.writeError(ctx, http.StatusBadRequest, errUnknownSyncEntity)
		return
	}
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	missing := make([]string, 0)
	for _, r := range req.Refs {
		if !present[r] {
			missing = append(missing, r)
		}
	}
	server.writeJSON(ctx, http.StatusOK, envelope{"missing": missing})
}
