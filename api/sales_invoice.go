package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type salesInvoiceItemRequest struct {
	ProductID int64 `json:"product_id" binding:"required"`
	UnitPrice int64 `json:"unit_price" binding:"required"`
	LineTotal int64 `json:"line_total" binding:"required"`
	Discount  int16 `json:"discount"`
	Quantity  int64 `json:"quantity" binding:"required"`
}

type createSalesInvoiceRequest struct {
	CashboxID        int64 `json:"cashbox_id" binding:"required"`
	CashboxAccountID int64 `json:"cashbox_account_id" binding:"required"`
	ShiftID          int64 `json:"shift_id" binding:"required"`
	InventoryID      int64 `json:"inventory_id" binding:"required"`
	Year             int32 `json:"year" binding:"required"`
	ClientID         int64 `json:"client_id" binding:"required"`
	Discount         int16 `json:"discount"`
	GrandTotal       int64 `json:"grand_total" binding:"required"`
	Subtotal         int64 `json:"sub_total" binding:"required"`
	DiscountedTotal  int64 `json:"discounted_total" binding:"required"`
	Items            []salesInvoiceItemRequest `json:"items" binding:"required,min=1,dive"`
	// TODO: we should see about this
	PriceListID *int64 `json:"price_list_id"`
}

func (server *Server) createSalesInvoice(ctx *gin.Context) {
	var req createSalesInvoiceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var priceListID sql.NullInt64
	if req.PriceListID != nil {
		priceListID = sql.NullInt64{Int64: *req.PriceListID, Valid: true}
	}

	items := make([]db.SalesInvoiceItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, db.SalesInvoiceItem{
			ProductID: item.ProductID,
			UnitPrice: item.UnitPrice,
			LineTotal: item.LineTotal,
			Discount:  item.Discount,
			Quantity:  item.Quantity,
		})
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
	}

	salesInvoice, err := server.store.SalesInvoiceTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
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
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	invoice, err := server.store.GetInvoice(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, invoice)
}

type listSalesInvoiceRequest struct {
	PageSize int32 `form:"page_size,default=10" binding:"min=5,max=10"`
	PageID   int32 `form:"page_id,default=0" binding:"min=0"`
}

func (server *Server) listSalesInvoices(ctx *gin.Context) {
	var req listSalesInvoiceRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListInvoicesParams{
		Limit:  req.PageSize,
		Offset: req.PageID,
	}
	invoices, err := server.store.ListInvoices(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	total, err := server.store.CountInvoices(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"sales_invoices": invoices, "total": total})
}
