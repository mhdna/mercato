package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type listLowStockProductsRequest struct {
	Search   string `form:"search"`
	Status   string `form:"status" binding:"omitempty,oneof=out_of_stock low_stock"`
	PageSize int32  `form:"page_size,default=20" binding:"min=1,max=100"`
	PageID   int32  `form:"page_id,default=0" binding:"min=0"`
}

// listLowStockProducts backs the Alerts page: every active product whose
// total on-hand (summed across every SKU and inventory) is at or below its
// effective threshold (its own override, else the global default). Status
// narrows to one bucket ("out_of_stock" or "low_stock"); omitted keeps both.
func (server *Server) listLowStockProducts(ctx *gin.Context) {
	var req listLowStockProductsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	var status sql.NullString
	if req.Status != "" {
		status = sql.NullString{String: req.Status, Valid: true}
	}

	respondList(server, ctx, "products",
		func() ([]db.ListLowStockProductsRow, error) {
			return server.store.ListLowStockProducts(ctx, db.ListLowStockProductsParams{
				Status:     status,
				Search:     req.Search,
				PageLimit:  req.PageSize,
				PageOffset: req.PageID,
			})
		},
		func() (int64, error) {
			return server.store.CountLowStockProducts(ctx, db.CountLowStockProductsParams{
				Status: status,
				Search: req.Search,
			})
		},
	)
}

type listIgnoredLowStockProductsRequest struct {
	Search   string `form:"search"`
	PageSize int32  `form:"page_size,default=20" binding:"min=1,max=100"`
	PageID   int32  `form:"page_id,default=0" binding:"min=0"`
}

// listIgnoredLowStockProducts lists products that have opted out of
// low-stock alerting entirely, so a muted product can still be found and
// un-muted later.
func (server *Server) listIgnoredLowStockProducts(ctx *gin.Context) {
	var req listIgnoredLowStockProductsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	respondList(server, ctx, "products",
		func() ([]db.ListIgnoredLowStockProductsRow, error) {
			return server.store.ListIgnoredLowStockProducts(ctx, db.ListIgnoredLowStockProductsParams{
				Search:     req.Search,
				PageLimit:  req.PageSize,
				PageOffset: req.PageID,
			})
		},
		func() (int64, error) {
			return server.store.CountIgnoredLowStockProducts(ctx, req.Search)
		},
	)
}

// getLowStockSummary backs the Alerts page's summary cards.
func (server *Server) getLowStockSummary(ctx *gin.Context) {
	summary, err := server.store.GetLowStockSummary(ctx)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, summary)
}

type setProductLowStockThresholdRequest struct {
	Threshold *int64 `json:"threshold"`
}

// setProductLowStockThreshold sets (or, with a null body, clears) one
// product's override threshold.
func (server *Server) setProductLowStockThreshold(ctx *gin.Context) {
	var uri getProductRequest
	if err := ctx.ShouldBindUri(&uri); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	var req setProductLowStockThresholdRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	var threshold sql.NullInt64
	if req.Threshold != nil {
		threshold = sql.NullInt64{Int64: *req.Threshold, Valid: true}
	}

	if err := server.store.UpdateProductLowStockThreshold(ctx, db.UpdateProductLowStockThresholdParams{
		ID:                uri.ID,
		LowStockThreshold: threshold,
	}); err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "threshold updated"})
}

type setProductLowStockAlertsEnabledRequest struct {
	Enabled bool `json:"enabled"`
}

// setProductLowStockAlertsEnabled mutes ("don't care if it's out of
// stock") or unmutes one product for low-stock alerting. A muted product
// is excluded from /low_stock and its summary regardless of on-hand.
func (server *Server) setProductLowStockAlertsEnabled(ctx *gin.Context) {
	var uri getProductRequest
	if err := ctx.ShouldBindUri(&uri); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	var req setProductLowStockAlertsEnabledRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	if err := server.store.SetProductLowStockAlertsEnabled(ctx, db.SetProductLowStockAlertsEnabledParams{
		ID:                    uri.ID,
		LowStockAlertsEnabled: req.Enabled,
	}); err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "updated"})
}

type setDefaultLowStockThresholdRequest struct {
	Threshold int64 `json:"threshold" binding:"min=0"`
}

// setDefaultLowStockThreshold changes the app-wide fallback threshold used
// by every product without its own override.
func (server *Server) setDefaultLowStockThreshold(ctx *gin.Context) {
	var req setDefaultLowStockThresholdRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	settings, err := server.store.SetDefaultLowStockThreshold(ctx, req.Threshold)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"settings": settings})
}
