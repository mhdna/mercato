package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type createBranchTargetSeriesRequest struct {
	TargetAmount int64 `json:"target_amount" binding:"required,min=1"`
	// StartDay is the day-of-month a period starts on (1-31, clamped in
	// short months -- see clampDay in branch_target_series.go). E.g. 1 for
	// plain calendar months, 4 for a "4th through the 3rd" billing cycle.
	StartDay int32 `json:"start_day" binding:"required,min=1,max=31"`
	// IntervalCount is how many months each period spans -- 1 for
	// monthly, 3 for quarterly, etc.
	IntervalCount int32  `json:"interval_count" binding:"required,min=1"`
	Color         string `json:"color"`
}

func (server *Server) createBranchTargetSeries(ctx *gin.Context) {
	var uri branchIDRequest
	if err := ctx.ShouldBindUri(&uri); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	var req createBranchTargetSeriesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	if req.Color != "" && !hexColorPattern.MatchString(req.Color) {
		server.writeError(ctx, http.StatusBadRequest, errInvalidBranchTargetColor)
		return
	}

	color := req.Color
	if color == "" {
		assigned, err := assignBranchTargetColor(ctx, server.store, uri.ID)
		if err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
		color = assigned
	}

	series, err := server.store.CreateBranchTargetSeries(ctx, db.CreateBranchTargetSeriesParams{
		BranchID:      uri.ID,
		TargetAmount:  req.TargetAmount,
		Color:         color,
		StartDay:      req.StartDay,
		IntervalCount: req.IntervalCount,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	// Fire the current period immediately rather than waiting for the
	// next hourly scheduler tick -- an admin creating a series expects to
	// see it show up in the targets list right away.
	fireDueBranchTargetSeries(ctx, server.store, server.branchHub)

	ctx.JSON(http.StatusOK, series)
}

func (server *Server) listBranchTargetSeries(ctx *gin.Context) {
	var uri branchIDRequest
	if err := ctx.ShouldBindUri(&uri); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	series, err := server.store.ListBranchTargetSeriesForBranch(ctx, uri.ID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, envelope{"series": series})
}

type setBranchTargetSeriesActiveRequest struct {
	ID     int64 `uri:"id" binding:"required,min=1"`
	Active bool  `json:"active"`
}

// setBranchTargetSeriesActive pauses/resumes a series without touching its
// schedule -- deleting it would lose start_day's cycle position, same
// reasoning as setRecurringExpenseActive. Pausing stops new periods from
// being generated; already-generated ones keep syncing and showing
// progress like any other branch target.
func (server *Server) setBranchTargetSeriesActive(ctx *gin.Context) {
	var req setBranchTargetSeriesActiveRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	series, err := server.store.SetBranchTargetSeriesActive(ctx, db.SetBranchTargetSeriesActiveParams{
		ID:     req.ID,
		Active: req.Active,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, series)
}

type updateBranchTargetSeriesRequest struct {
	ID            int64  `uri:"id" binding:"required,min=1"`
	TargetAmount  int64  `json:"target_amount" binding:"required,min=1"`
	StartDay      int32  `json:"start_day" binding:"required,min=1,max=31"`
	IntervalCount int32  `json:"interval_count" binding:"required,min=1"`
	// Optional; blank keeps the series' current color rather than clearing it.
	Color string `json:"color"`
}

// updateBranchTargetSeries only affects periods generated after this call
// -- see the note on CreateGeneratedBranchTarget. There's no in-place
// "edit progress" concept since periods are independent snapshots.
func (server *Server) updateBranchTargetSeries(ctx *gin.Context) {
	var req updateBranchTargetSeriesRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	if req.Color != "" && !hexColorPattern.MatchString(req.Color) {
		server.writeError(ctx, http.StatusBadRequest, errInvalidBranchTargetColor)
		return
	}

	color := req.Color
	if color == "" {
		existing, err := server.store.GetBranchTargetSeries(ctx, req.ID)
		if err != nil {
			if err == sql.ErrNoRows {
				server.writeError(ctx, http.StatusNotFound, err)
				return
			}
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
		color = existing.Color
	}

	series, err := server.store.UpdateBranchTargetSeries(ctx, db.UpdateBranchTargetSeriesParams{
		ID:            req.ID,
		TargetAmount:  req.TargetAmount,
		Color:         color,
		StartDay:      req.StartDay,
		IntervalCount: req.IntervalCount,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, series)
}

type deleteBranchTargetSeriesRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// deleteBranchTargetSeries removes the series and deactivates every period
// it ever generated (see DeleteBranchTargetSeriesTx) -- deleting the whole
// recurring schedule at once, not just one already-generated occurrence.
func (server *Server) deleteBranchTargetSeries(ctx *gin.Context) {
	var req deleteBranchTargetSeriesRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	series, err := server.store.DeleteBranchTargetSeriesTx(ctx, req.ID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	server.branchHub.notify(series.BranchID, branchWSMessage{Type: "target_updated"})
	ctx.JSON(http.StatusOK, envelope{"deleted": true})
}
