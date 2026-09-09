package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

// stockHealthStatus classifies overall inventory health into the same
// traffic-light the sidebar dot uses. Deliberately independent of the
// low-stock alerts (too little stock, per product): this is the opposite
// axis -- too much stock, or a sell-through rate that would empty the
// shelves too fast. Heuristic, not a hard SLA:
//   - critical: inventory would sell through in under 3 days at the
//     current rate, or more than a fifth of the active catalog is dead
//     stock (zero sales in 90 days despite being on hand).
//   - warning: would sell through within two weeks, or more than a tenth
//     of the catalog is dead stock or overstocked.
//   - healthy: otherwise.
func stockHealthStatus(daysOfInventory *float64, deadStock, overstocked, totalActive int64) string {
	deadRatio := ratio(deadStock, totalActive)
	slowRatio := ratio(deadStock+overstocked, totalActive)

	if daysOfInventory != nil && *daysOfInventory < 3 {
		return "critical"
	}
	if deadRatio > 0.2 {
		return "critical"
	}
	if daysOfInventory != nil && *daysOfInventory < 14 {
		return "warning"
	}
	if slowRatio > 0.1 {
		return "warning"
	}
	return "healthy"
}

func ratio(part, whole int64) float64 {
	if whole == 0 {
		return 0
	}
	return float64(part) / float64(whole)
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
// sidebar status dot: overall days-of-inventory, inventory value, dead
// stock / overstock counts, and a traffic-light status. The per-inventory
// breakdown and the slow-mover list are separate, paginated endpoints
// (listStockHealthByInventory, listSlowMovingProducts) -- neither ships as
// one unbounded list.
func (server *Server) getStockHealth(ctx *gin.Context) {
	overview, err := server.store.GetStockHealthOverview(ctx)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	slowMovers, err := server.store.GetSlowMoverSummary(ctx)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	days := asFloat64Ptr(overview.DaysOfInventory)
	status := stockHealthStatus(days, slowMovers.DeadStockCount, slowMovers.OverstockedCount, slowMovers.TotalActiveProducts)

	ctx.JSON(http.StatusOK, gin.H{
		"status":                status,
		"days_of_inventory":     days,
		"total_units":           overview.TotalUnits,
		"inventory_value":       overview.InventoryValue,
		"units_sold_30d":        overview.UnitsSold30d,
		"dead_stock_count":      slowMovers.DeadStockCount,
		"overstocked_count":     slowMovers.OverstockedCount,
		"total_active_products": slowMovers.TotalActiveProducts,
	})
}

type listSlowMovingProductsRequest struct {
	Search   string `form:"search"`
	Category string `form:"category" binding:"omitempty,oneof=dead overstocked"`
	PageSize int32  `form:"page_size,default=20" binding:"min=1,max=100"`
	PageID   int32  `form:"page_id,default=0" binding:"min=0"`
}

// listSlowMovingProducts backs the "Recommended For Sale" table: active
// products carrying stock that either never sold in the last 90 days
// ("dead") or would take over 90 days to sell through at their current
// pace ("overstocked"). Category narrows to one bucket; omitted keeps both.
func (server *Server) listSlowMovingProducts(ctx *gin.Context) {
	var req listSlowMovingProductsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	var category sql.NullString
	if req.Category != "" {
		category = sql.NullString{String: req.Category, Valid: true}
	}

	respondList(server, ctx, "products",
		func() ([]db.ListSlowMovingProductsRow, error) {
			return server.store.ListSlowMovingProducts(ctx, db.ListSlowMovingProductsParams{
				Category:   category,
				Search:     req.Search,
				PageLimit:  req.PageSize,
				PageOffset: req.PageID,
			})
		},
		func() (int64, error) {
			return server.store.CountSlowMovingProducts(ctx, db.CountSlowMovingProductsParams{
				Category: category,
				Search:   req.Search,
			})
		},
	)
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
