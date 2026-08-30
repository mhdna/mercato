package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type invoiceDetailsRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// getInvoiceDetails is the shared "everything about this invoice" view for
// both kinds of kashi's own local invoices (sales_invoices/return_invoices
// are thin tag tables over the same invoices row) -- client, salesperson
// (if one was recorded), line items, the settlement breakdown per cashbox
// account, and the loyalty points delta applied. Used by sales-invoices.vue
// (and, once a details action exists there, a return-invoice view too).
func (server *Server) getInvoiceDetails(ctx *gin.Context) {
	var req invoiceDetailsRequest
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

	client, err := server.store.GetClient(ctx, invoice.ClientID)
	if err != nil && err != sql.ErrNoRows {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	var salesperson *db.Salesperson
	if invoice.SalespersonID.Valid {
		sp, err := server.store.GetSalesperson(ctx, invoice.SalespersonID.Int64)
		if err == nil {
			salesperson = &sp
		} else if err != sql.ErrNoRows {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
	}

	items, err := server.store.ListInvoiceProductsByInvoice(ctx, invoice.ID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	payments, err := server.store.ListInvoicePaymentsByInvoice(ctx, invoice.ID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	kind := "sales"
	var relatedSalesInvoiceID *int64
	if returnInvoice, err := server.store.GetReturnInvoice(ctx, invoice.ID); err == nil {
		kind = "return"
		relatedSalesInvoiceID = &returnInvoice.SalesInvoiceID
	} else if err != sql.ErrNoRows {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{
		"invoice":                  invoice,
		"kind":                     kind,
		"related_sales_invoice_id": relatedSalesInvoiceID,
		"client":                   client,
		"salesperson":              salesperson,
		"items":                    items,
		"payments":                 payments,
	})
}
