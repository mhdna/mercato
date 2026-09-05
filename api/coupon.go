package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

// couponCategoryID maps the optional category_id from a request (0 = none)
// to the nullable column value.
func couponCategoryID(id int64) sql.NullInt64 {
	if id > 0 {
		return sql.NullInt64{Int64: id, Valid: true}
	}
	return sql.NullInt64{}
}

type createCouponRequest struct {
	Code         string `json:"code" binding:"required"`
	Status       string `json:"status" binding:"required"`
	DiscountType string `json:"discount_type" binding:"required"`
	Reason       string `json:"reason" binding:"required"`
	ClientID     int64  `json:"client_id" binding:"required"`
	ValidUntil   string `json:"valid_until" binding:"required"`
	CategoryID   int64  `json:"category_id"`
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
		CategoryID:   couponCategoryID(req.CategoryID),
	}

	coupon, err := server.store.CreateCoupon(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, coupon)
}

type updateCouponURI struct {
	Code string `uri:"code" binding:"required"`
}

type updateCouponRequest struct {
	Status       string `json:"status" binding:"required"`
	DiscountType string `json:"discount_type" binding:"required"`
	Reason       string `json:"reason" binding:"required"`
	ClientID     int64  `json:"client_id" binding:"required"`
	ValidUntil   string `json:"valid_until" binding:"required"`
	CategoryID   int64  `json:"category_id"`
}

func (server *Server) updateCoupon(ctx *gin.Context) {
	var uri updateCouponURI
	if err := ctx.ShouldBindUri(&uri); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	var req updateCouponRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	validUntil, err := time.Parse(time.RFC3339, req.ValidUntil)
	if err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	coupon, err := server.store.UpdateCoupon(ctx, db.UpdateCouponParams{
		Code:         uri.Code,
		Status:       db.CouponStatus(req.Status),
		DiscountType: db.DiscountType(req.DiscountType),
		Reason:       req.Reason,
		ClientID:     req.ClientID,
		ValidUntil:   validUntil,
		CategoryID:   couponCategoryID(req.CategoryID),
	})
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
	PageSize   int32  `form:"page_size,default=10" binding:"min=5,max=100"`
	PageID     int32  `form:"page_id,default=0" binding:"min=0"`
	Search     string `form:"search"`
	Status     string `form:"status" binding:"omitempty,oneof=active inactive"`
	CategoryID int64  `form:"category_id"`
}

func (server *Server) listCoupons(ctx *gin.Context) {
	var req listCouponsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	var status db.NullCouponStatus
	if req.Status != "" {
		status = db.NullCouponStatus{CouponStatus: db.CouponStatus(req.Status), Valid: true}
	}
	categoryID := couponCategoryID(req.CategoryID)

	respondList(server, ctx, "coupons",
		func() ([]db.Coupon, error) {
			return server.store.ListCoupons(ctx, db.ListCouponsParams{
				Search:     req.Search,
				Status:     status,
				CategoryID: categoryID,
				PageSize:   req.PageSize,
				PageOffset: req.PageID,
			})
		},
		func() (int64, error) {
			return server.store.CountCoupons(ctx, db.CountCouponsParams{
				Search:     req.Search,
				Status:     status,
				CategoryID: categoryID,
			})
		},
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

type bulkDeleteCouponsRequest struct {
	Codes []string `json:"codes" binding:"required,min=1,max=500,dive,min=1"`
}

func (server *Server) bulkDeleteCoupons(ctx *gin.Context) {
	var req bulkDeleteCouponsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	rows, err := server.store.DeleteCoupons(ctx, req.Codes)
	server.bulkRespond(ctx, rows, err)
}
