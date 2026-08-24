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

type listCouponsRequest struct {
	PageSize int32 `form:"page_size,default=10" binding:"min=5,max=10"`
	PageID   int32 `form:"page_id,default=0" binding:"min=0"`
}

func (server *Server) listCoupons(ctx *gin.Context) {
	var req listCouponsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.ListCouponsParams{
		Limit:  req.PageSize,
		Offset: req.PageID,
	}
	coupons, err := server.store.ListCoupons(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountCoupons(ctx)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"coupons": coupons, "total": total})
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
