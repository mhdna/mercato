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

var errRelatedClientRefRequired = errors.New("related_client_ref is required for a return or exchange invoice")

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
	// Barcode is the SKU's globally-unique barcode; it's how kashi resolves
	// a branch line to a central product_variants row so the branch's store
	// inventory can be decremented. Optional -- an older kashi-pos build
	// that hasn't been updated omits it, and such a line is recorded but
	// doesn't move central stock.
	Barcode   string `json:"barcode"`
	UnitPrice int64  `json:"unit_price"`
	LineTotal int64  `json:"line_total"`
	Discount  int16  `json:"discount"`
	Quantity  int64  `json:"quantity" binding:"required"`
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

	// Both best-effort display data, not synced/shared entities -- see the
	// matching comment on branchInvoicePaymentPayload in kashi-pos's
	// sync_outbox.go. Neither is required: an older kashi-pos build that
	// hasn't been updated yet simply omits them.
	SalespersonName string                        `json:"salesperson_name"`
	Payments        []branchInvoicePaymentRequest `json:"payments"`

	// Historical marks a one-time backfill of a pre-sync sale: it records in
	// full but moves no central stock (see CreateBranchInvoiceTx). Absent /
	// false on every live checkout.
	Historical bool `json:"historical"`
}

type branchInvoicePaymentRequest struct {
	AccountName string `json:"account_name" binding:"required"`
	Amount      int64  `json:"amount"`
}

func (server *Server) createBranchInvoice(ctx *gin.Context, kind string) {
	branchID := ctx.MustGet(branchIDKey).(int64)

	var req branchInvoiceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	if (kind == "return" || kind == "exchange") && req.RelatedClientRef == "" {
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
			Barcode:         item.Barcode,
			UnitPrice:       item.UnitPrice,
			LineTotal:       item.LineTotal,
			Discount:        item.Discount,
			Quantity:        item.Quantity,
		})
	}

	payments := make([]db.BranchInvoicePaymentParams, 0, len(req.Payments))
	for _, payment := range req.Payments {
		payments = append(payments, db.BranchInvoicePaymentParams{
			AccountName: payment.AccountName,
			Amount:      payment.Amount,
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
			SalespersonName:        req.SalespersonName,
			Historical:             req.Historical,
		},
		Items:    items,
		Payments: payments,
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

	// Only the genuinely-new-insert path reaches here (the two lookups above
	// return early on "already exists"), so this never double-notifies
	// admin clients for a retried/idempotent replay.
	//
	// Invoices don't carry their own currency_code (unlike branch expenses),
	// so the toast displays the system's single default currency. A failed
	// lookup just means an empty currency code on the toast -- informational
	// only, never worth failing the checkout over.
	currencyCode := ""
	if currency, err := server.store.GetDefaultCurrency(ctx); err == nil {
		currencyCode = currency.Code
	}
	// GrandTotal is already the net signed cash effect the branch reported
	// (positive for sales, negative for a refund return, and for an
	// exchange whatever kashi-pos's own netDifference computed -- positive,
	// negative, or zero) -- see the ListDailyIncome comment in
	// db/query/branch_invoice.sql. Forcing a second negation here for
	// "return" used to double-flip it back to positive.
	server.adminHub.broadcastAll(adminWSMessage{
		Type:         "branch_invoice_created",
		BranchID:     branchID,
		Kind:         kind,
		Amount:       req.GrandTotal,
		CurrencyCode: currencyCode,
		Label:        req.BranchInvoiceCode,
	})

	server.writeJSON(ctx, http.StatusOK, envelope{"invoice": result.Invoice, "items": result.Items, "payments": result.Payments})
}

func (server *Server) createBranchSalesInvoice(ctx *gin.Context) {
	server.createBranchInvoice(ctx, "sales")
}

func (server *Server) createBranchReturnInvoice(ctx *gin.Context) {
	server.createBranchInvoice(ctx, "return")
}

func (server *Server) createBranchExchangeInvoice(ctx *gin.Context) {
	server.createBranchInvoice(ctx, "exchange")
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

// isForeignKeyViolation reports whether err is a Postgres foreign-key
// constraint error -- used to turn "still referenced elsewhere" deletes
// (e.g. a loan category still used by a loan) into a 409 instead of a
// generic 500.
func isForeignKeyViolation(err error) bool {
	pqErr, ok := err.(*pq.Error)
	return ok && pqErr.Code.Name() == "foreign_key_violation"
}
