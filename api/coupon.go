package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type createCouponRequest struct {
	Code         string `json:"code" binding:"required"`
	Status       string `json:"status" binding:"required"`
	DiscountType string `json:"discount_type" binding:"required"`
	Reason       string `json:"reason" binding:"required"`
	ClientID     int64  `json:"client_id" binding:"required"`
	ValidUntil   string `json:"valid_until" binding:"required"`
}

func (server *Server) createCoupon(ctx *gin.Context) {
	var req createCouponRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	validUntil, err := time.Parse(time.RFC3339, req.ValidUntil)
	if err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.CreateCouponParams{
		Code:         req.Code,
		Status:       db.CouponStatus(req.Status),
		DiscountType: db.DiscountType(req.DiscountType),
		Reason:       req.Reason,
		ClientID:     req.ClientID,
		ValidUntil:   validUntil,
	}

	coupon, err := server.store.CreateCoupon(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, coupon)
}

type getCouponRequest struct {
	Code string `uri:"code" binding:"required"`
}

func (server *Server) getCoupon(ctx *gin.Context) {
	var req getCouponRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	coupon, err := server.store.GetCoupon(ctx, req.Code)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, coupon)
}

func (server *Server) listCoupons(ctx *gin.Context) {
	q, ok := server.bindListPageQuery(ctx)
	if !ok {
		return
	}

	respondList(server, ctx, "coupons",
		func() ([]db.Coupon, error) {
			return server.store.ListCoupons(ctx, db.ListCouponsParams{
				Search: q.Search, PageSize: q.PageSize, PageOffset: q.PageID,
			})
		},
		func() (int64, error) { return server.store.CountCoupons(ctx, q.Search) },
	)
}

type deactivateCouponRequest struct {
	Code string `uri:"code" binding:"required"`
}

func (server *Server) deactivateCoupon(ctx *gin.Context) {
	var req deactivateCouponRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	err := server.store.DeactivateCoupon(ctx, req.Code)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "coupon deactivated"})
}
