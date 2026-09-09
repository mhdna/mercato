package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

// stockHealthStatus classifies overall inventory health into the same
// traffic-light the sidebar dot uses. Heuristic, not a hard SLA:
//   - critical: inventory would run out in under 3 days at the current
//     sell-through rate, or more than 10% of monitored products are
//     out of stock.
//   - warning: any product is out of stock, low, or inventory would run
//     out within two weeks.
//   - healthy: otherwise.
func stockHealthStatus(daysOfInventory *float64, outOfStock, lowStock, totalMonitored int64) string {
	if daysOfInventory != nil && *daysOfInventory < 3 {
		return "critical"
	}
	if totalMonitored > 0 && float64(outOfStock)/float64(totalMonitored) > 0.1 {
		return "critical"
	}
	if outOfStock > 0 || lowStock > 0 {
		return "warning"
	}
	if daysOfInventory != nil && *daysOfInventory < 14 {
		return "warning"
	}
	return "healthy"
}

func asFloat64Ptr(v interface{}) *float64 {
	switch n := v.(type) {
	case float64:
		return &n
	case float32:
		f := float64(n)
		return &f
	default:
		return nil
	}
}

// getStockHealth backs the Stock Health page's summary cards and the
// sidebar status dot: overall days-of-inventory, value, and a
// traffic-light status derived from those plus the existing low-stock
// summary. The per-inventory breakdown is a separate, paginated endpoint
// (listStockHealthByInventory) -- an install can have hundreds of
// inventories (one per branch), so it never ships as one unbounded list.
func (server *Server) getStockHealth(ctx *gin.Context) {
	overview, err := server.store.GetStockHealthOverview(ctx)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	summary, err := server.store.GetLowStockSummary(ctx)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	days := asFloat64Ptr(overview.DaysOfInventory)
	status := stockHealthStatus(days, summary.OutOfStockCount, summary.LowStockCount, summary.TotalProducts)

	ctx.JSON(http.StatusOK, gin.H{
		"status":             status,
		"days_of_inventory":  days,
		"total_units":        overview.TotalUnits,
		"inventory_value":    overview.InventoryValue,
		"units_sold_30d":     overview.UnitsSold30d,
		"out_of_stock_count": summary.OutOfStockCount,
		"low_stock_count":    summary.LowStockCount,
		"total_products":     summary.TotalProducts,
	})
}

type listStockHealthByInventoryRequest struct {
	Search   string `form:"search"`
	PageSize int32  `form:"page_size,default=20" binding:"min=1,max=100"`
	PageID   int32  `form:"page_id,default=0" binding:"min=0"`
}

// listStockHealthByInventory backs the "By Location" table on the Stock
// Health page: the same days-of-inventory/value breakdown, one row per
// inventory, paginated and searchable like every other list page.
func (server *Server) listStockHealthByInventory(ctx *gin.Context) {
	var req listStockHealthByInventoryRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	respondList(server, ctx, "inventories",
		func() ([]db.ListStockHealthByInventoryRow, error) {
			return server.store.ListStockHealthByInventory(ctx, db.ListStockHealthByInventoryParams{
				Search:     req.Search,
				PageLimit:  req.PageSize,
				PageOffset: req.PageID,
			})
		},
		func() (int64, error) {
			return server.store.CountStockHealthByInventory(ctx, req.Search)
		},
	)
}
