package api

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	db "github.com/mhdna/kashi/db/sqlc"
)

var errRelatedClientRefRequired = errors.New("related_client_ref is required for a return invoice")

// branchHealth lets a branch confirm its API key works and see how kashi
// identifies it, before wiring up anything that actually moves data.
func (server *Server) branchHealth(ctx *gin.Context) {
	branchID := ctx.MustGet(branchIDKey).(int64)

	server.writeJSON(ctx, http.StatusOK, envelope{"branch_id": branchID})
}

// UnitPrice/LineTotal deliberately have no "required" validation tag: Go's
// validator treats the zero value of a plain numeric field as "missing"
// under required, but 0 is a legitimate amount here -- e.g. a line with a
// 100% discount has a real, valid line_total of 0. Quantity keeps
// "required" since a zero-quantity line never makes sense (returns use
// negative quantities, which pass required fine).
type branchInvoiceItemRequest struct {
	BranchProductID int64 `json:"branch_product_id" binding:"required"`
	UnitPrice       int64 `json:"unit_price"`
	LineTotal       int64 `json:"line_total"`
	Discount        int16 `json:"discount"`
	Quantity        int64 `json:"quantity" binding:"required"`
}

// branchInvoiceRequest is deliberately NOT the same shape as
// createSalesInvoiceRequest/createReturnInvoiceRequest (the human-operator
// endpoints). Those reference kashi's own cashbox/shift/client/inventory
// rows by ID and let kashi compute loyalty points and cashbox balances
// itself. Neither is true here:
//
//   - BranchCashboxAccountID/BranchShiftID/BranchInventoryID/
//     BranchClientID/items[].BranchProductID are the branch's own local
//     SQLite IDs — meaningless to kashi's schema until branch-scoped
//     identity sync exists (a later phase). They're stored as plain
//     traceability data, not foreign keys. (kashi-pos has no separate
//     "cashbox register" entity of its own — its invoices.cashbox_id is
//     actually a cashbox_account id — so there's no parallel
//     BranchCashboxID field here either.)
//   - LoyaltyPointsDelta is whatever kashi-pos already computed and
//     applied locally at checkout. kashi records it as reported; it must
//     never recompute this from GrandTotal using its own formula, since
//     the two codebases' formulas have already been found to disagree.
type branchInvoiceRequest struct {
	ClientRef              string `json:"client_ref" binding:"required"`
	BranchInvoiceCode      string `json:"branch_invoice_code" binding:"required"`
	BranchCashboxAccountID int64  `json:"branch_cashbox_account_id" binding:"required"`
	BranchShiftID          int64  `json:"branch_shift_id" binding:"required"`
	BranchInventoryID      int64  `json:"branch_inventory_id" binding:"required"`
	BranchClientID         *int64 `json:"branch_client_id"`
	Discount               int16  `json:"discount"`
	// GrandTotal/Subtotal/DiscountedTotal have no "required" tag for the
	// same reason as branchInvoiceItemRequest.LineTotal above: 0 is a
	// legitimate amount (a fully-discounted invoice), not a missing field.
	GrandTotal         int64                      `json:"grand_total"`
	Subtotal           int64                      `json:"sub_total"`
	DiscountedTotal    int64                      `json:"discounted_total"`
	LoyaltyPointsDelta int64                      `json:"loyalty_points_delta"`
	OccurredAt         time.Time                  `json:"occurred_at" binding:"required"`
	Items              []branchInvoiceItemRequest `json:"items" binding:"required,min=1,dive"`

	// Required for returns only: the client_ref (within this same branch)
	// of the sale being returned/exchanged against.
	RelatedClientRef string `json:"related_client_ref"`
}

func (server *Server) createBranchInvoice(ctx *gin.Context, kind string) {
	branchID := ctx.MustGet(branchIDKey).(int64)

	var req branchInvoiceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	if kind == "return" && req.RelatedClientRef == "" {
		server.writeError(ctx, http.StatusBadRequest, errRelatedClientRefRequired)
		return
	}

	if existing, err := server.store.GetBranchInvoiceByClientRef(ctx, db.GetBranchInvoiceByClientRefParams{
		BranchID:  branchID,
		ClientRef: req.ClientRef,
	}); err == nil {
		server.writeJSON(ctx, http.StatusOK, envelope{"invoice": existing})
		return
	} else if err != sql.ErrNoRows {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	var branchClientID sql.NullInt64
	if req.BranchClientID != nil {
		branchClientID = sql.NullInt64{Int64: *req.BranchClientID, Valid: true}
	}
	var relatedRef sql.NullString
	if req.RelatedClientRef != "" {
		relatedRef = sql.NullString{String: req.RelatedClientRef, Valid: true}
	}

	items := make([]db.BranchInvoiceItemParams, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, db.BranchInvoiceItemParams{
			BranchProductID: item.BranchProductID,
			UnitPrice:       item.UnitPrice,
			LineTotal:       item.LineTotal,
			Discount:        item.Discount,
			Quantity:        item.Quantity,
		})
	}

	result, err := server.store.CreateBranchInvoiceTx(ctx, db.CreateBranchInvoiceTxParams{
		CreateBranchInvoiceParams: db.CreateBranchInvoiceParams{
			BranchID:               branchID,
			ClientRef:              req.ClientRef,
			Kind:                   kind,
			BranchInvoiceCode:      req.BranchInvoiceCode,
			BranchCashboxAccountID: req.BranchCashboxAccountID,
			BranchShiftID:          req.BranchShiftID,
			BranchInventoryID:      req.BranchInventoryID,
			BranchClientID:         branchClientID,
			RelatedClientRef:       relatedRef,
			Discount:               req.Discount,
			Subtotal:               req.Subtotal,
			DiscountedTotal:        req.DiscountedTotal,
			GrandTotal:             req.GrandTotal,
			LoyaltyPointsDelta:     req.LoyaltyPointsDelta,
			OccurredAt:             req.OccurredAt,
		},
		Items: items,
	})
	if err != nil {
		if isUniqueViolation(err) {
			if existing, lookupErr := server.store.GetBranchInvoiceByClientRef(ctx, db.GetBranchInvoiceByClientRefParams{
				BranchID:  branchID,
				ClientRef: req.ClientRef,
			}); lookupErr == nil {
				server.writeJSON(ctx, http.StatusOK, envelope{"invoice": existing})
				return
			}
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"invoice": result.Invoice, "items": result.Items})
}

func (server *Server) createBranchSalesInvoice(ctx *gin.Context) {
	server.createBranchInvoice(ctx, "sales")
}

func (server *Server) createBranchReturnInvoice(ctx *gin.Context) {
	server.createBranchInvoice(ctx, "return")
}

// isUniqueViolation reports whether err is a Postgres unique-constraint
// error — used to catch the race where two retries of the same client_ref
// land concurrently: the check-then-create above isn't atomic on its own,
// so the unique constraint on (branch_id, client_ref) is the real
// correctness backstop, and this lets us treat "lost the race" the same as
// "already exists" rather than surfacing it as an error.
func isUniqueViolation(err error) bool {
	pqErr, ok := err.(*pq.Error)
	return ok && pqErr.Code.Name() == "unique_violation"
}
