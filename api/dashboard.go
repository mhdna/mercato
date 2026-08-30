package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

// dashboard.go backs the home dashboard's Overview tab. Everything here is
// read-only aggregation over a [date_from, date_to] day range and an
// optional branch, with a `scope` that selects branch-synced data, kashi's
// own admin data, or both combined -- see db/query/dashboard.sql for how
// the two sources are defined and unioned.

type dashboardFilter struct {
	Scope    string `form:"scope,default=all" binding:"oneof=all branch admin"`
	BranchID int64  `form:"branch_id"`
	DateFrom string `form:"date_from"`
	DateTo   string `form:"date_to"`
}

type dashboardListFilter struct {
	dashboardFilter
	PageSize int32 `form:"page_size,default=10" binding:"min=5,max=100"`
	// ServerSideTable.vue sends page_id already multiplied out to a row
	// offset ((page-1)*itemsPerPage), so it's used as OFFSET directly --
	// same as listBranchInvoices.
	PageID int32 `form:"page_id,default=0" binding:"min=0"`
}

// resolveRange turns the optional YYYY-MM-DD strings into a half-open
// [from, to) instant pair: `to` is the start of the day *after* date_to so
// the whole of that last day is included. Defaults to the current calendar
// month when either bound is missing/unparseable.
func (f dashboardFilter) resolveRange() (from, to time.Time) {
	now := time.Now().UTC()
	from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	to = from.AddDate(0, 1, 0)

	if parsed, err := time.Parse("2006-01-02", f.DateFrom); err == nil {
		from = parsed.UTC()
	}
	if parsed, err := time.Parse("2006-01-02", f.DateTo); err == nil {
		to = parsed.UTC().AddDate(0, 0, 1)
	}
	return from, to
}

func (f dashboardFilter) branchArg() sql.NullInt64 {
	if f.BranchID > 0 {
		return sql.NullInt64{Int64: f.BranchID, Valid: true}
	}
	return sql.NullInt64{}
}

// effectiveScope is the requested scope after accounting for a selected
// branch: admin/central records aren't attributable to any branch, so
// picking one drops them entirely (an "all branches" view is the only place
// they belong -- same rule as ListClientsBySpending). This keeps the cards,
// charts and every table telling the same story for a given filter set.
func (f dashboardFilter) effectiveScope() string {
	if f.BranchID > 0 && f.Scope != "admin" {
		return "branch"
	}
	return f.Scope
}

func (f dashboardFilter) wantsBranch() bool { return f.effectiveScope() != "admin" }
func (f dashboardFilter) wantsAdmin() bool  { return f.effectiveScope() != "branch" }

type dashboardSeriesPoint struct {
	Day       string `json:"day"`
	Revenue   int64  `json:"revenue"`
	Expenses  int64  `json:"expenses"`
	Purchases int64  `json:"purchases"`
}

type dashboardTargetRow struct {
	BranchName   string `json:"branch_name"`
	TargetAmount int64  `json:"target_amount"`
	Revenue      int64  `json:"revenue"`
	Pct          int64  `json:"pct"`
	Color        string `json:"color"`
}

type dashboardDueLoanRow struct {
	Description string `json:"description"`
	BranchName  string `json:"branch_name"`
	Origin      string `json:"origin"`
	Outstanding int64  `json:"outstanding"`
}

// getDashboardSummary powers the five Overview cards and the four trend
// charts in one round trip.
func (server *Server) getDashboardSummary(ctx *gin.Context) {
	var req dashboardFilter
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	from, to := req.resolveRange()
	branch := req.branchArg()

	var itemsSold, newClients, revenue, expenses, purchases, invoices int64

	// day (YYYY-MM-DD) -> [revenue, expenses, purchases], zero-filled for
	// every day in range after both sources are folded in.
	revByDay := map[string]int64{}
	expByDay := map[string]int64{}
	purchByDay := map[string]int64{}

	if req.wantsBranch() {
		summary, err := server.store.DashboardBranchSummary(ctx, db.DashboardBranchSummaryParams{
			DateFrom: from, DateTo: to, BranchID: branch,
		})
		if err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
		itemsSold += summary.ItemsSold
		revenue += summary.Revenue
		invoices += summary.Invoices

		exp, err := server.store.DashboardBranchExpensesTotal(ctx, db.DashboardBranchExpensesTotalParams{
			DateFrom: from, DateTo: to, BranchID: branch,
		})
		if err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
		expenses += exp

		clients, err := server.store.DashboardBranchNewClients(ctx, db.DashboardBranchNewClientsParams{
			DateFrom: from, DateTo: to, BranchID: branch,
		})
		if err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
		newClients += clients

		revSeries, err := server.store.DashboardBranchRevenueSeries(ctx, db.DashboardBranchRevenueSeriesParams{
			DateFrom: from, DateTo: to, BranchID: branch,
		})
		if err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
		for _, row := range revSeries {
			revByDay[row.Day.Format("2006-01-02")] += row.Total
		}

		expSeries, err := server.store.DashboardBranchExpenseSeries(ctx, db.DashboardBranchExpenseSeriesParams{
			DateFrom: from, DateTo: to, BranchID: branch,
		})
		if err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
		for _, row := range expSeries {
			expByDay[row.Day.Format("2006-01-02")] += row.Total
		}
	}

	// effectiveScope already drops admin when a branch is selected (unless
	// admin scope was picked explicitly).
	if req.wantsAdmin() {
		summary, err := server.store.DashboardAdminSummary(ctx, db.DashboardAdminSummaryParams{
			DateFrom: from, DateTo: to,
		})
		if err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
		itemsSold += summary.ItemsSold
		revenue += summary.Revenue
		invoices += summary.Invoices

		exp, err := server.store.DashboardAdminExpensesTotal(ctx, db.DashboardAdminExpensesTotalParams{
			DateFrom: from, DateTo: to,
		})
		if err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
		expenses += exp

		clients, err := server.store.DashboardAdminNewClients(ctx, db.DashboardAdminNewClientsParams{
			DateFrom: from, DateTo: to,
		})
		if err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
		newClients += clients

		revSeries, err := server.store.DashboardAdminRevenueSeries(ctx, db.DashboardAdminRevenueSeriesParams{
			DateFrom: from, DateTo: to,
		})
		if err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
		for _, row := range revSeries {
			revByDay[row.Day.Format("2006-01-02")] += row.Total
		}

		expSeries, err := server.store.DashboardAdminExpenseSeries(ctx, db.DashboardAdminExpenseSeriesParams{
			DateFrom: from, DateTo: to,
		})
		if err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
		for _, row := range expSeries {
			expByDay[row.Day.Format("2006-01-02")] += row.Total
		}

		// Purchases (cost of goods bought) are admin-only -- there is no
		// branch purchase concept -- so they only ever fold in here.
		purch, err := server.store.DashboardPurchasesTotal(ctx, db.DashboardPurchasesTotalParams{
			DateFrom: from, DateTo: to,
		})
		if err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
		purchases += purch

		purchSeries, err := server.store.DashboardPurchaseSeries(ctx, db.DashboardPurchaseSeriesParams{
			DateFrom: from, DateTo: to,
		})
		if err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
		for _, row := range purchSeries {
			purchByDay[row.Day.Format("2006-01-02")] += row.Total
		}
	}

	series := make([]dashboardSeriesPoint, 0)
	for day := from; day.Before(to); day = day.AddDate(0, 0, 1) {
		key := day.Format("2006-01-02")
		series = append(series, dashboardSeriesPoint{
			Day:       key,
			Revenue:   revByDay[key],
			Expenses:  expByDay[key],
			Purchases: purchByDay[key],
		})
	}

	// Due loans: outstanding balance (not range-bound), plus the biggest few
	// for the dashboard list. Both reflect effectiveScope so a selected
	// branch only sees its own branch loans.
	dueLoans, err := server.store.DashboardDueLoans(ctx, db.DashboardDueLoansParams{
		Scope:    req.effectiveScope(),
		BranchID: branch,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	dueLoanRows, err := server.store.DashboardTopDueLoans(ctx, db.DashboardTopDueLoansParams{
		Scope:    req.effectiveScope(),
		BranchID: branch,
		TopN:     5,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	dueLoansTop := make([]dashboardDueLoanRow, 0, len(dueLoanRows))
	for _, row := range dueLoanRows {
		dueLoansTop = append(dueLoansTop, dashboardDueLoanRow{
			Description: row.Description,
			BranchName:  row.BranchName,
			Origin:      row.Origin,
			Outstanding: row.Outstanding,
		})
	}

	// Targets: one row (progress bar) per branch_target active today, plus
	// the overall attainment percent. Empty / null under admin scope, since
	// branch_targets are branch-only.
	var targetsPct *int64
	targets := make([]dashboardTargetRow, 0)
	if req.wantsBranch() {
		rows, err := server.store.DashboardTargetList(ctx, db.DashboardTargetListParams{
			AsOf:     time.Now().UTC(),
			BranchID: branch,
		})
		if err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
		var totalTarget, totalRevenue int64
		for _, row := range rows {
			pct := int64(0)
			if row.TargetAmount > 0 {
				pct = row.Revenue * 100 / row.TargetAmount
			}
			targets = append(targets, dashboardTargetRow{
				BranchName:   row.BranchName,
				TargetAmount: row.TargetAmount,
				Revenue:      row.Revenue,
				Pct:          pct,
				Color:        row.Color,
			})
			totalTarget += row.TargetAmount
			totalRevenue += row.Revenue
		}
		if totalTarget > 0 {
			pct := totalRevenue * 100 / totalTarget
			targetsPct = &pct
		}
	}

	server.writeJSON(ctx, http.StatusOK, envelope{
		"items_sold":    itemsSold,
		"new_clients":   newClients,
		"revenue":       revenue,
		"expenses":      expenses,
		"purchases":     purchases,
		"invoices":      invoices,
		"due_loans":     dueLoans,
		"due_loans_top": dueLoansTop,
		"targets_pct":   targetsPct,
		"targets":       targets,
		"series":        series,
	})
}

func (server *Server) listDashboardSales(ctx *gin.Context) {
	var req dashboardListFilter
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	from, to := req.resolveRange()
	branch := req.branchArg()

	sales, err := server.store.DashboardSalesList(ctx, db.DashboardSalesListParams{
		Limit: req.PageSize, Offset: req.PageID,
		Scope: req.effectiveScope(), DateFrom: from, DateTo: to, BranchID: branch,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountDashboardSales(ctx, db.CountDashboardSalesParams{
		Scope: req.effectiveScope(), DateFrom: from, DateTo: to, BranchID: branch,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"sales": sales, "total": total})
}

func (server *Server) listDashboardPurchases(ctx *gin.Context) {
	var req dashboardListFilter
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	from, to := req.resolveRange()

	purchases, err := server.store.DashboardPurchasesList(ctx, db.DashboardPurchasesListParams{
		Limit: req.PageSize, Offset: req.PageID,
		Scope: req.effectiveScope(), DateFrom: from, DateTo: to,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountDashboardPurchases(ctx, db.CountDashboardPurchasesParams{
		Scope: req.effectiveScope(), DateFrom: from, DateTo: to,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"purchases": purchases, "total": total})
}

func (server *Server) listDashboardExpenses(ctx *gin.Context) {
	var req dashboardListFilter
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	from, to := req.resolveRange()
	branch := req.branchArg()

	expenses, err := server.store.DashboardExpensesList(ctx, db.DashboardExpensesListParams{
		Limit: req.PageSize, Offset: req.PageID,
		Scope: req.effectiveScope(), DateFrom: from, DateTo: to, BranchID: branch,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountDashboardExpenses(ctx, db.CountDashboardExpensesParams{
		Scope: req.effectiveScope(), DateFrom: from, DateTo: to, BranchID: branch,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"expenses": expenses, "total": total})
}

// listDashboardActivities returns a single, globally ordered page across
// every persisted branch-sync activity type. Keeping the union in SQL makes
// the count and pagination accurate without the UI fan-out fetching each
// underlying resource independently.
func (server *Server) listDashboardActivities(ctx *gin.Context) {
	var req dashboardListFilter
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	from, to := req.resolveRange()

	activities, err := server.store.DashboardActivitiesList(ctx, db.DashboardActivitiesListParams{
		Scope: req.effectiveScope(), DateFrom: from, DateTo: to,
		BranchID: req.branchArg(), PageSize: req.PageSize, PageID: req.PageID,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	var total int64
	if len(activities) > 0 {
		total = activities[0].TotalCount
	}
	server.writeJSON(ctx, http.StatusOK, envelope{"activities": activities, "total": total})
}
