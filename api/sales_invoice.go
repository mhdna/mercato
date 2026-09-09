package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type salesInvoiceItemRequest struct {
	VariantID int64 `json:"variant_id" binding:"required"`
	UnitPrice int64 `json:"unit_price" binding:"required"`
	LineTotal int64 `json:"line_total" binding:"required"`
	Discount  int16 `json:"discount"`
	Quantity  int64 `json:"quantity" binding:"required"`
}

type createSalesInvoiceRequest struct {
	CashboxID        int64                     `json:"cashbox_id" binding:"required"`
	CashboxAccountID int64                     `json:"cashbox_account_id" binding:"required"`
	ShiftID          int64                     `json:"shift_id" binding:"required"`
	InventoryID      int64                     `json:"inventory_id" binding:"required"`
	Year             int32                     `json:"year" binding:"required"`
	ClientID         int64                     `json:"client_id" binding:"required"`
	Discount         int16                     `json:"discount"`
	GrandTotal       int64                     `json:"grand_total" binding:"required"`
	Subtotal         int64                     `json:"sub_total" binding:"required"`
	DiscountedTotal  int64                     `json:"discounted_total" binding:"required"`
	Items            []salesInvoiceItemRequest `json:"items" binding:"required,min=1,dive"`
	// TODO: we should see about this
	PriceListID   *int64 `json:"price_list_id"`
	InvoiceTypeID *int64 `json:"invoice_type_id"`
	SalespersonID *int64 `json:"salesperson_id"`
}

func (server *Server) createSalesInvoice(ctx *gin.Context) {
	var req createSalesInvoiceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	var priceListID sql.NullInt64
	if req.PriceListID != nil {
		priceListID = sql.NullInt64{Int64: *req.PriceListID, Valid: true}
	}

	var invoiceTypeID int64
	if req.InvoiceTypeID != nil {
		invoiceTypeID = *req.InvoiceTypeID
	} else {
		defaultInvoiceType, err := server.store.GetDefaultInvoiceType(ctx)
		if err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
		invoiceTypeID = defaultInvoiceType.ID
	}

	items := make([]db.SalesInvoiceItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, db.SalesInvoiceItem{
			VariantID: item.VariantID,
			UnitPrice: item.UnitPrice,
			LineTotal: item.LineTotal,
			Discount:  item.Discount,
			Quantity:  item.Quantity,
		})
	}

	var salespersonID sql.NullInt64
	if req.SalespersonID != nil {
		salespersonID = sql.NullInt64{Int64: *req.SalespersonID, Valid: true}
	}

	arg := db.SalesInvoiceTxParams{
		CashboxID:        req.CashboxID,
		CashboxAccountID: req.CashboxAccountID,
		ShiftID:          req.ShiftID,
		InventoryID:      req.InventoryID,
		Year:             req.Year,
		ClientID:         req.ClientID,
		Discount:         req.Discount,
		GrandTotal:       req.GrandTotal,
		SubTotal:         req.Subtotal,
		DiscountedTotal:  req.DiscountedTotal,
		Items:            items,
		PriceListID:      priceListID,
		InvoiceTypeID:    invoiceTypeID,
		SalespersonID:    salespersonID,
		CreatedBy:        actorID(ctx),
	}

	salesInvoice, err := server.store.SalesInvoiceTx(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, salesInvoice)
}

type getSalesInvoiceRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getSalesInvoice(ctx *gin.Context) {
	var req getSalesInvoiceRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	invoice, err := server.store.GetInvoice(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}

		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, invoice)
}

type listSalesInvoiceRequest struct {
	PageSize int32  `form:"page_size,default=14" binding:"min=5,max=100"`
	PageID   int32  `form:"page_id,default=0" binding:"min=0"`
	Search   string `form:"search"`
}

func (server *Server) listSalesInvoices(ctx *gin.Context) {
	var req listSalesInvoiceRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.ListInvoicesPageParams{
		Search:     req.Search,
		PageSize:   req.PageSize,
		PageOffset: req.PageID,
	}
	invoices, err := server.store.ListInvoicesPage(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountInvoicesFiltered(ctx, req.Search)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"sales_invoices": invoices, "total": total})
}
