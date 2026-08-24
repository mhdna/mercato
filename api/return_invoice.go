package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type returnInvoiceItemRequest struct {
	ProductID int64 `json:"product_id" binding:"required"`
	UnitPrice int64 `json:"unit_price" binding:"required"`
	LineTotal int64 `json:"line_total" binding:"required"`
	Discount  int16 `json:"discount"`
	Quantity  int64 `json:"quantity" binding:"required"`
}

type createReturnInvoiceRequest struct {
	CashboxID        int64                      `json:"cashbox_id" binding:"required"`
	CashboxAccountID int64                      `json:"cashbox_account_id" binding:"required"`
	ShiftID          int64                      `json:"shift_id" binding:"required"`
	InventoryID      int64                      `json:"inventory_id" binding:"required"`
	Year             int32                      `json:"year" binding:"required"`
	ClientID         int64                      `json:"client_id" binding:"required"`
	SalesInvoiceID   int64                      `json:"sales_invoice_id" binding:"required"`
	Discount         int16                      `json:"discount"`
	GrandTotal       int64                      `json:"grand_total" binding:"required"`
	Subtotal         int64                      `json:"sub_total" binding:"required"`
	DiscountedTotal  int64                      `json:"discounted_total" binding:"required"`
	Items            []returnInvoiceItemRequest `json:"items" binding:"required,min=1,dive"`
	PriceListID      *int64                     `json:"price_list_id"`
}

func (server *Server) createReturnInvoice(ctx *gin.Context) {
	var req createReturnInvoiceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	var priceListID sql.NullInt64
	if req.PriceListID != nil {
		priceListID = sql.NullInt64{Int64: *req.PriceListID, Valid: true}
	}

	items := make([]db.ReturnInvoiceItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, db.ReturnInvoiceItem{
			ProductID: item.ProductID,
			UnitPrice: item.UnitPrice,
			LineTotal: item.LineTotal,
			Discount:  item.Discount,
			Quantity:  item.Quantity,
		})
	}

	arg := db.ReturnInvoiceTxParams{
		CashboxID:        req.CashboxID,
		CashboxAccountID: req.CashboxAccountID,
		ShiftID:          req.ShiftID,
		InventoryID:      req.InventoryID,
		Year:             req.Year,
		ClientID:         req.ClientID,
		SalesInvoiceID:   req.SalesInvoiceID,
		Discount:         req.Discount,
		GrandTotal:       req.GrandTotal,
		SubTotal:         req.Subtotal,
		DiscountedTotal:  req.DiscountedTotal,
		Items:            items,
		PriceListID:      priceListID,
	}

	returnInvoice, err := server.store.ReturnInvoiceTx(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, returnInvoice)
}

type getReturnInvoiceRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getReturnInvoice(ctx *gin.Context) {
	var req getReturnInvoiceRequest
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

type listReturnInvoiceRequest struct {
	PageSize int32 `form:"page_size,default=10" binding:"min=5,max=10"`
	PageID   int32 `form:"page_id,default=0" binding:"min=0"`
}

func (server *Server) listReturnInvoices(ctx *gin.Context) {
	var req listReturnInvoiceRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.ListInvoicesParams{
		Limit:  req.PageSize,
		Offset: req.PageID,
	}
	invoices, err := server.store.ListInvoices(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, invoices)
}
