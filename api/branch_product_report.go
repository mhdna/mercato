package api

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// branchProductReportItem is one flat POS product row the till couldn't
// resolve to a central variant by barcode. It carries everything the till
// knows so an admin can adopt it into the catalog from the conflict screen.
type branchProductReportItem struct {
	Barcode      string `json:"barcode" binding:"required"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Price        int64  `json:"price"`
	Color        string `json:"color"`
	Size         string `json:"size"`
	Brand        string `json:"brand"`
	Kind         string `json:"kind"`
	Season       string `json:"season"`
	Year         string `json:"year"`
	CurrencyCode string `json:"currency_code"`
}

type branchProductReportRequest struct {
	Products []branchProductReportItem `json:"products" binding:"required,min=1,dive"`
}

// reportBranchProducts stages one 'unknown_product' conflict per barcode the
// branch reported that kashi has no product_variants row for. A barcode that
// *does* resolve is silently skipped -- the branch's own reconcile should
// already have filtered those, but re-reporting a now-known one is harmless.
// Idempotent on (branch_id, 'product', barcode).
func (server *Server) reportBranchProducts(ctx *gin.Context) {
	branchID := ctx.MustGet(branchIDKey).(int64)

	var req branchProductReportRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	staged := 0
	for _, p := range req.Products {
		if _, err := server.store.GetProductVariantByBarcode(ctx, p.Barcode); err == nil {
			continue // already in the catalog
		}
		payload, _ := json.Marshal(p)
		if server.stageBranchSyncConflict(ctx, branchID, "product", p.Barcode, "unknown_product", payload, nullInt64None()) {
			staged++
		}
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"staged": staged})
}
