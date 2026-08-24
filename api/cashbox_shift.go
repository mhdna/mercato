package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type createShiftRequest struct {
	CashboxID int64 `json:"cashbox_id" binding:"required,min=1"`
}

func (server *Server) createShift(ctx *gin.Context) {
	var req createShiftRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	cashboxID := req.CashboxID

	shift, err := server.store.CreateShift(ctx, cashboxID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, shift)
}

type getShiftRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getShift(ctx *gin.Context) {
	var req getShiftRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	shift, err := server.store.GetShift(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}

		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, shift)
}

type listShifts struct {
	PageSize int32 `form:"page_size,default=10" binding:"min=5,max=10"`
	PageID   int32 `form:"page_id,default=0" binding:"min=0"`
}

func (server *Server) listShifts(ctx *gin.Context) {
	var req listShifts
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.ListShiftsParams{
		Limit:  req.PageSize,
		Offset: req.PageID,
	}
	shifts, err := server.store.ListShifts(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountShifts(ctx)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"shifts": shifts, "total": total})
}

type CloseShiftRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) CloseShift(ctx *gin.Context) {
	var req CloseShiftRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	shiftID := req.ID

	err := server.store.CloseShift(ctx, shiftID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusBadRequest, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "closed shift"})
}
