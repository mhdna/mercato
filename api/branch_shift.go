package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

// branchShiftCloseRequest is the exact JSON body POSTed to
// /branch/shift_closes by kashi-pos (see branchShiftClosePayload in that
// repo's sync_outbox.go, which must stay in sync with this shape).
// BranchShiftID is the branch's own local shifts.id, meaningless to kashi's
// schema -- stored as opaque traceability data, not a foreign key, same
// reasoning as branchInvoiceRequest in branch_sync.go. Each variance is
// derived below as counted - expected; a branch-reported variance is never
// trusted.
type branchShiftCloseRequest struct {
	ShiftRef      string `json:"shift_ref" binding:"required"`
	BranchShiftID int64  `json:"branch_shift_id" binding:"required"`
	// opened_at/closed_at have no "required" tag: an older kashi-pos build,
	// or one that couldn't parse its own stored timestamp, simply omits
	// them, and a zero time is stored as NULL rather than rejected.
	OpenedAt          time.Time `json:"opened_at"`
	ClosedAt          time.Time `json:"closed_at"`
	ClosingPersonName string    `json:"closing_person_name"`
	OpeningFloatUsd   int64     `json:"opening_float_usd"`
	OpeningFloatLbp   int64     `json:"opening_float_lbp"`
	// Expected/Counted amounts have no "required" tag for the same reason as
	// branchInvoiceRequest.GrandTotal: 0 is a legitimate amount (an empty
	// shift), not a missing field.
	ExpectedUsd   int64     `json:"expected_usd"`
	ExpectedLbp   int64     `json:"expected_lbp"`
	ExpectedVisa  int64     `json:"expected_visa"`
	ExpectedWhish int64     `json:"expected_whish"`
	CountedUsd    int64     `json:"counted_usd"`
	CountedLbp    int64     `json:"counted_lbp"`
	CountedVisa   int64     `json:"counted_visa"`
	CountedWhish  int64     `json:"counted_whish"`
	OccurredAt    time.Time `json:"occurred_at" binding:"required"`
}

func nullTimeIfSet(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t, Valid: true}
}

// createBranchShiftClose follows the same idempotent-insert pattern as
// createBranchExpense in branch_expense.go: look up by (branch_id,
// client_ref) first so a retried sync_outbox entry is a no-op, and fall
// back to the same lookup on a unique-violation race.
func (server *Server) createBranchShiftClose(ctx *gin.Context) {
	branchID := ctx.MustGet(branchIDKey).(int64)

	var req branchShiftCloseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	if existing, err := server.store.GetBranchShiftByClientRef(ctx, db.GetBranchShiftByClientRefParams{
		BranchID:  branchID,
		ClientRef: req.ShiftRef,
	}); err == nil {
		server.writeJSON(ctx, http.StatusOK, envelope{"shift": existing})
		return
	} else if err != sql.ErrNoRows {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	shift, err := server.store.CreateBranchShift(ctx, db.CreateBranchShiftParams{
		BranchID:          branchID,
		ClientRef:         req.ShiftRef,
		BranchShiftID:     req.BranchShiftID,
		OpenedAt:          nullTimeIfSet(req.OpenedAt),
		ClosedAt:          nullTimeIfSet(req.ClosedAt),
		ClosingPersonName: req.ClosingPersonName,
		OpeningFloatUsd:   req.OpeningFloatUsd,
		OpeningFloatLbp:   req.OpeningFloatLbp,
		ExpectedUsd:       req.ExpectedUsd,
		ExpectedLbp:       req.ExpectedLbp,
		ExpectedVisa:      req.ExpectedVisa,
		ExpectedWhish:     req.ExpectedWhish,
		CountedUsd:        req.CountedUsd,
		CountedLbp:        req.CountedLbp,
		CountedVisa:       req.CountedVisa,
		CountedWhish:      req.CountedWhish,
		VarianceUsd:       req.CountedUsd - req.ExpectedUsd,
		VarianceLbp:       req.CountedLbp - req.ExpectedLbp,
		VarianceVisa:      req.CountedVisa - req.ExpectedVisa,
		VarianceWhish:     req.CountedWhish - req.ExpectedWhish,
		OccurredAt:        req.OccurredAt,
	})
	if err != nil {
		if isUniqueViolation(err) {
			if existing, lookupErr := server.store.GetBranchShiftByClientRef(ctx, db.GetBranchShiftByClientRefParams{
				BranchID:  branchID,
				ClientRef: req.ShiftRef,
			}); lookupErr == nil {
				server.writeJSON(ctx, http.StatusOK, envelope{"shift": existing})
				return
			}
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	// Only the genuinely-new-insert path reaches here (the lookup above
	// returns early on "already exists"), same convention as
	// createBranchInvoice -- so a retried outbox entry never re-notifies.
	// The toast carries the USD-drawer variance, signed (negative is a
	// shortfall) -- the one figure a manager reacts to at close time.
	currencyCode := ""
	if currency, err := server.store.GetDefaultCurrency(ctx); err == nil {
		currencyCode = currency.Code
	}
	server.adminHub.broadcastAll(adminWSMessage{
		Type:         "branch_shift_closed",
		BranchID:     branchID,
		Amount:       shift.VarianceUsd,
		CurrencyCode: currencyCode,
		Label:        req.ClosingPersonName,
	})

	server.writeJSON(ctx, http.StatusOK, envelope{"shift": shift})
}

type listBranchShiftsRequest struct {
	PageSize int32 `form:"page_size,default=10" binding:"min=5,max=100"`
	PageID   int32 `form:"page_id,default=0" binding:"min=0"`
	BranchID int64 `form:"branch_id"`
}

// listBranchShifts is the admin-facing counterpart to createBranchShiftClose,
// same pagination shape as listBranchExpenses.
func (server *Server) listBranchShifts(ctx *gin.Context) {
	var req listBranchShiftsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	var branchID sql.NullInt64
	if req.BranchID > 0 {
		branchID = sql.NullInt64{Int64: req.BranchID, Valid: true}
	}

	shifts, err := server.store.ListBranchShifts(ctx, db.ListBranchShiftsParams{
		Limit:    req.PageSize,
		Offset:   req.PageID,
		BranchID: branchID,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountBranchShifts(ctx, branchID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"branch_shifts": shifts, "total": total})
}
