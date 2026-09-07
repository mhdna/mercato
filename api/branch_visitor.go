package api

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

// branchVisitorEventRequest is the exact JSON body POSTed to
// /branch/visitor_events by kashi-pos (see branchVisitorEventPayload in
// that repo's sync_outbox.go, which must stay in sync with this shape).
// One press of the nav-drawer people counter = one request. Like
// branch_shifts, kashi keeps these as an append-only record of what the
// till reported, not an integration into a central footfall entity --
// there is none.
//
// Day is the "YYYY-MM-DD" local-calendar day the till bucketed the press
// under; it's stored verbatim so the per-day rollup matches the till's own
// without kashi re-deriving a calendar day from OccurredAt's timezone.
type branchVisitorEventRequest struct {
	Ref        string    `json:"ref" binding:"required"`
	Day        string    `json:"day" binding:"required"`
	Direction  string    `json:"direction" binding:"required,oneof=in out"`
	OccurredAt time.Time `json:"occurred_at" binding:"required"`
}

// createBranchVisitorEvent follows the same idempotent-insert pattern as
// createBranchShiftClose: look up by (branch_id, client_ref) first so a
// retried sync_outbox entry is a no-op, and fall back to the same lookup
// on a unique-violation race.
//
// A genuinely-new press broadcasts a "branch_visitor_event" on the admin
// hub (Kind = "in"/"out", Label = the local day) so the SyncCard ticker /
// toast reflects it and the Visitors views refetch without a manual
// reload. It's a higher-frequency event than the others, so whether it
// actually surfaces as a toast is left to the admin's existing
// activity-display setting, same as every other branch activity type.
func (server *Server) createBranchVisitorEvent(ctx *gin.Context) {
	branchID := ctx.MustGet(branchIDKey).(int64)

	var req branchVisitorEventRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	if existing, err := server.store.GetBranchVisitorEventByClientRef(ctx, db.GetBranchVisitorEventByClientRefParams{
		BranchID:  branchID,
		ClientRef: req.Ref,
	}); err == nil {
		server.writeJSON(ctx, http.StatusOK, envelope{"visitor_event": existing})
		return
	} else if err != sql.ErrNoRows {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	event, err := server.store.CreateBranchVisitorEvent(ctx, db.CreateBranchVisitorEventParams{
		BranchID:   branchID,
		ClientRef:  req.Ref,
		Day:        req.Day,
		Direction:  req.Direction,
		OccurredAt: req.OccurredAt,
	})
	if err != nil {
		if isUniqueViolation(err) {
			if existing, lookupErr := server.store.GetBranchVisitorEventByClientRef(ctx, db.GetBranchVisitorEventByClientRefParams{
				BranchID:  branchID,
				ClientRef: req.Ref,
			}); lookupErr == nil {
				server.writeJSON(ctx, http.StatusOK, envelope{"visitor_event": existing})
				return
			}
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	// Only the genuinely-new-insert path reaches here (the lookups above
	// return early on "already exists"), so a retried outbox entry never
	// re-notifies -- same convention as createBranchShiftClose. Amount
	// carries the running gross tally for this branch/day/direction so the
	// toast can read "customer in -- 42 in today" (a plain count, like
	// branch_attendance_events, not cents).
	soFar, err := server.store.CountBranchVisitorDayByDirection(ctx, db.CountBranchVisitorDayByDirectionParams{
		BranchID:  branchID,
		Day:       event.Day,
		Direction: event.Direction,
	})
	if err != nil {
		soFar = 0 // count is display-only; don't fail the write over it
	}
	server.adminHub.broadcastAll(adminWSMessage{
		Type:     "branch_visitor_event",
		BranchID: branchID,
		Kind:     event.Direction,
		Amount:   soFar,
		Label:    event.Day,
	})

	server.writeJSON(ctx, http.StatusOK, envelope{"visitor_event": event})
}

type listBranchVisitorDaysRequest struct {
	PageSize int32 `form:"page_size,default=10" binding:"min=5,max=100"`
	PageID   int32 `form:"page_id,default=0" binding:"min=0"`
	BranchID int64 `form:"branch_id"`
}

// listBranchVisitorDays is the admin-facing counterpart to
// createBranchVisitorEvent: one row per (branch, local day) with gross
// customers-in / customers-out, newest day first. Same pagination shape as
// listBranchShifts; page_id is an offset, not a page number.
func (server *Server) listBranchVisitorDays(ctx *gin.Context) {
	var req listBranchVisitorDaysRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	var branchID sql.NullInt64
	if req.BranchID > 0 {
		branchID = sql.NullInt64{Int64: req.BranchID, Valid: true}
	}

	days, err := server.store.ListBranchVisitorDays(ctx, db.ListBranchVisitorDaysParams{
		Limit:    req.PageSize,
		Offset:   req.PageID,
		BranchID: branchID,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountBranchVisitorDays(ctx, branchID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"visitor_days": days, "total": total})
}

type listBranchVisitorStatsRequest struct {
	From     string `form:"from" binding:"required,datetime=2006-01-02"`
	To       string `form:"to" binding:"required,datetime=2006-01-02"`
	BranchID int64  `form:"branch_id"`
}

// listBranchVisitorStats backs the Footfall page: every (branch, day)
// rollup inside an inclusive date range, unpaginated. The client turns
// this into month-over-month bars, branch comparisons, weekday profiles
// and a month x branch table -- keeping the aggregation client-side means
// one endpoint serves every cut without a query per chart.
func (server *Server) listBranchVisitorStats(ctx *gin.Context) {
	var req listBranchVisitorStatsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	if req.From > req.To {
		server.writeError(ctx, http.StatusBadRequest, errors.New("from date is after to date"))
		return
	}

	var branchID sql.NullInt64
	if req.BranchID > 0 {
		branchID = sql.NullInt64{Int64: req.BranchID, Valid: true}
	}

	days, err := server.store.ListBranchVisitorDaysRange(ctx, db.ListBranchVisitorDaysRangeParams{
		FromDay:  req.From,
		ToDay:    req.To,
		BranchID: branchID,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"visitor_days": days})
}
