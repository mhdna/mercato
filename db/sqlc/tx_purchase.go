package db

import (
	"context"
	"database/sql"
	"errors"
)

// PurchaseItemParams is one line on a draft purchase invoice. A line is
// either a product variant (variant_id set) or an asset (asset_id set).
type PurchaseItemParams struct {
	VariantID sql.NullInt64 `json:"variant_id"`
	AssetID   sql.NullInt64 `json:"asset_id"`
	Quantity  int64         `json:"quantity"`
	UnitPrice int64         `json:"unit_price"`
}

type CreatePurchaseTxParams struct {
	SupplierID   int64                `json:"supplier_id"`
	InventoryID  int64                `json:"inventory_id"`
	Code         string               `json:"code"`
	CurrencyCode string               `json:"currency_code"`
	PurchasedAt  sql.NullTime         `json:"purchased_at"`
	Note         string               `json:"note"`
	Items        []PurchaseItemParams `json:"items"`
}

type CreatePurchaseTxResult struct {
	Purchase Purchase       `json:"purchase"`
	Items    []PurchaseItem `json:"items"`
}

// CreatePurchaseTx writes a draft purchase invoice and its lines together,
// and stamps the header subtotal/grand_total from the lines. It does not
// touch stock -- that happens on receive.
func (store *SQLStore) CreatePurchaseTx(ctx context.Context, arg CreatePurchaseTxParams) (CreatePurchaseTxResult, error) {
	var result CreatePurchaseTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		purchasedAt := arg.PurchasedAt
		if !purchasedAt.Valid {
			purchasedAt = nowNullTime()
		}

		purchase, err := q.CreatePurchase(ctx, CreatePurchaseParams{
			SupplierID:   arg.SupplierID,
			InventoryID:  sql.NullInt64{Int64: arg.InventoryID, Valid: arg.InventoryID != 0},
			Code:         arg.Code,
			CurrencyCode: arg.CurrencyCode,
			PurchasedAt:  purchasedAt.Time,
			Note:         arg.Note,
		})
		if err != nil {
			return err
		}

		var subtotal int64
		items := make([]PurchaseItem, 0, len(arg.Items))
		for _, it := range arg.Items {
			created, err := q.AddPurchaseItem(ctx, AddPurchaseItemParams{
				PurchaseID:   sql.NullInt64{Int64: purchase.ID, Valid: true},
				VariantID:    it.VariantID,
				AssetID:      it.AssetID,
				Quantity:     it.Quantity,
				UnitPrice:    it.UnitPrice,
				CurrencyCode: arg.CurrencyCode,
			})
			if err != nil {
				return err
			}
			subtotal += it.UnitPrice * it.Quantity
			items = append(items, created)
		}

		if err := q.SetPurchaseTotals(ctx, SetPurchaseTotalsParams{
			ID:         purchase.ID,
			Subtotal:   subtotal,
			GrandTotal: subtotal,
		}); err != nil {
			return err
		}
		purchase.Subtotal = subtotal
		purchase.GrandTotal = subtotal

		result.Purchase = purchase
		result.Items = items
		return nil
	})

	return result, err
}

type PurchaseReceiveTxParams struct {
	PurchaseID int64         `json:"purchase_id"`
	CreatedBy  sql.NullInt64 `json:"created_by"`
}

type PurchaseReceiveTxResult struct {
	Purchase  Purchase        `json:"purchase"`
	Movements []StockMovement `json:"movements"`
}

// PurchaseReceiveTx moves a draft purchase to 'received': every product
// line lands in the purchase's destination inventory as a 'purchase' stock
// movement, rolls that SKU's per-location and global moving-average cost
// forward, and records the supplier's cost for the product. Idempotent --
// a purchase that is already received is returned unchanged.
func (store *SQLStore) PurchaseReceiveTx(ctx context.Context, arg PurchaseReceiveTxParams) (PurchaseReceiveTxResult, error) {
	var result PurchaseReceiveTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		purchase, err := q.GetPurchase(ctx, arg.PurchaseID)
		if err != nil {
			return err
		}
		if purchase.Status == PurchaseStatusReceived {
			result.Purchase = purchaseFromRow(purchase)
			return nil
		}
		if purchase.Status == PurchaseStatusCancelled {
			return errors.New("cannot receive a cancelled purchase")
		}
		if !purchase.InventoryID.Valid {
			return errors.New("purchase has no destination inventory")
		}
		inventoryID := purchase.InventoryID.Int64

		items, err := q.ListPurchaseItems(ctx, nullInt64(arg.PurchaseID))
		if err != nil {
			return err
		}

		for _, it := range items {
			if !it.VariantID.Valid {
				continue // asset lines don't enter variant stock
			}
			variantID := it.VariantID.Int64

			variant, err := q.GetProductVariant(ctx, variantID)
			if err != nil {
				return err
			}

			globalOnHand, err := q.GetVariantTotalOnHand(ctx, variantID)
			if err != nil {
				return err
			}
			if err := q.UpdateProductVariantAvgCost(ctx, UpdateProductVariantAvgCostParams{
				CurrentQty: globalOnHand,
				UnitCost:   it.UnitPrice,
				InQty:      it.Quantity,
				ID:         variantID,
			}); err != nil {
				return err
			}

			movement, err := q.applyStockMovement(ctx, applyStockMovementParams{
				InventoryID:   inventoryID,
				VariantID:     variantID,
				Quantity:      it.Quantity,
				Reason:        StockMovementReasonPurchase,
				ReferenceType: "purchase",
				ReferenceID:   purchase.ID,
				UnitCost:      sql.NullInt64{Int64: it.UnitPrice, Valid: true},
				CreatedBy:     arg.CreatedBy,
			})
			if err != nil {
				return err
			}
			result.Movements = append(result.Movements, movement)

			if err := q.recordSupplierCost(ctx, variant.ProductID, purchase.SupplierID, it.UnitPrice, purchase.CurrencyCode); err != nil {
				return err
			}
		}

		updated, err := q.SetPurchaseStatus(ctx, SetPurchaseStatusParams{
			Status:     PurchaseStatusReceived,
			ReceivedAt: nowNullTime(),
			ID:         purchase.ID,
		})
		if err != nil {
			return err
		}
		result.Purchase = updated
		return nil
	})

	return result, err
}

// recordSupplierCost keeps product_suppliers / product_supplier_costs in
// step with what a purchase actually paid, as a "last known cost from this
// supplier" reference. Best-effort: a duplicate (product_supplier_id,
// unit_cost) is silently ignored by the query's ON CONFLICT.
func (q *Queries) recordSupplierCost(ctx context.Context, productID, supplierID, unitCost int64, currencyCode string) error {
	ps, err := q.GetProductSupplier(ctx, GetProductSupplierParams{
		ProductID:  productID,
		SupplierID: supplierID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		ps, err = q.AddPurchasedProduct(ctx, AddPurchasedProductParams{
			ProductID:  productID,
			SupplierID: supplierID,
		})
	}
	if err != nil {
		return err
	}
	_, err = q.AddPurchasedProductCost(ctx, AddPurchasedProductCostParams{
		ProductSupplierID: ps.ID,
		UnitCost:          unitCost,
		CurrencyCode:      currencyCode,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil // ON CONFLICT DO NOTHING -> no row returned
	}
	return err
}

// purchaseFromRow narrows a GetPurchaseRow (header + joined names) back to
// a plain Purchase for the "already received" early return.
func purchaseFromRow(r GetPurchaseRow) Purchase {
	return Purchase{
		ID:           r.ID,
		SupplierID:   r.SupplierID,
		PurchasedAt:  r.PurchasedAt,
		InventoryID:  r.InventoryID,
		Code:         r.Code,
		CurrencyCode: r.CurrencyCode,
		Subtotal:     r.Subtotal,
		GrandTotal:   r.GrandTotal,
		Status:       r.Status,
		ReceivedAt:   r.ReceivedAt,
		Note:         r.Note,
	}
}
