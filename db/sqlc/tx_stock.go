package db

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// applyStockMovementParams describes one signed change to a SKU's on-hand
// stock in one inventory.
type applyStockMovementParams struct {
	InventoryID   int64
	VariantID     int64
	Quantity      int64 // signed: negative removes stock
	Reason        StockMovementReason
	ReferenceType string
	ReferenceID   int64 // 0 => no document (bare adjustment/count)
	UnitCost      sql.NullInt64
	Note          string
	CreatedBy     sql.NullInt64
}

// applyStockMovement appends a stock_movements ledger row (the source of
// truth) and rolls the cached inventory_stock on-hand forward. For an
// inbound movement that carries a unit cost (a purchase receipt, or the
// receiving leg of a transfer) it also rolls the per-location
// moving-average cost. It never blocks on negative on-hand -- the ledger
// stays authoritative and the UI surfaces negatives.
func (q *Queries) applyStockMovement(ctx context.Context, arg applyStockMovementParams) (StockMovement, error) {
	var refType sql.NullString
	if arg.ReferenceType != "" {
		refType = sql.NullString{String: arg.ReferenceType, Valid: true}
	}
	var refID sql.NullInt64
	if arg.ReferenceID != 0 {
		refID = sql.NullInt64{Int64: arg.ReferenceID, Valid: true}
	}

	movement, err := q.CreateStockMovement(ctx, CreateStockMovementParams{
		InventoryID:   arg.InventoryID,
		VariantID:     arg.VariantID,
		Quantity:      arg.Quantity,
		Reason:        arg.Reason,
		ReferenceType: refType,
		ReferenceID:   refID,
		UnitCost:      arg.UnitCost,
		Note:          arg.Note,
		CreatedBy:     arg.CreatedBy,
	})
	if err != nil {
		return StockMovement{}, err
	}

	if arg.UnitCost.Valid && arg.Quantity > 0 {
		if _, err := q.ReceiveInventoryStock(ctx, ReceiveInventoryStockParams{
			InventoryID: arg.InventoryID,
			VariantID:   arg.VariantID,
			Quantity:    arg.Quantity,
			UnitCost:    arg.UnitCost.Int64,
		}); err != nil {
			return StockMovement{}, err
		}
		return movement, nil
	}

	if _, err := q.AddInventoryStockQuantity(ctx, AddInventoryStockQuantityParams{
		InventoryID: arg.InventoryID,
		VariantID:   arg.VariantID,
		Quantity:    arg.Quantity,
	}); err != nil {
		return StockMovement{}, err
	}
	return movement, nil
}

// ---------------------------------------------------------------------------
// Stock adjustment (no supporting document)
// ---------------------------------------------------------------------------

type StockAdjustmentMode string

const (
	// StockAdjustmentModeDelta adds a signed quantity to on-hand.
	StockAdjustmentModeDelta StockAdjustmentMode = "delta"
	// StockAdjustmentModeCount sets on-hand to an absolute counted figure.
	StockAdjustmentModeCount StockAdjustmentMode = "count"
)

type StockAdjustmentTxParams struct {
	InventoryID int64               `json:"inventory_id"`
	VariantID   int64               `json:"variant_id"`
	Mode        StockAdjustmentMode `json:"mode"`
	Quantity    int64               `json:"quantity"` // delta: signed change; count: target on-hand
	Note        string              `json:"note"`
	CreatedBy   sql.NullInt64       `json:"created_by"`
}

type StockAdjustmentTxResult struct {
	Movement StockMovement `json:"movement"`
	OnHand   int64         `json:"on_hand"`
}

func (store *SQLStore) StockAdjustmentTx(ctx context.Context, arg StockAdjustmentTxParams) (StockAdjustmentTxResult, error) {
	var result StockAdjustmentTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		delta := arg.Quantity
		reason := StockMovementReasonAdjustment

		if arg.Mode == StockAdjustmentModeCount {
			reason = StockMovementReasonCount
			current := int64(0)
			existing, err := q.GetInventoryStock(ctx, GetInventoryStockParams{
				InventoryID: arg.InventoryID,
				VariantID:   arg.VariantID,
			})
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return err
			}
			if err == nil {
				current = existing.Quantity
			}
			delta = arg.Quantity - current
		} else if arg.Mode != StockAdjustmentModeDelta {
			return errors.New("invalid stock adjustment mode")
		}

		movement, err := q.applyStockMovement(ctx, applyStockMovementParams{
			InventoryID: arg.InventoryID,
			VariantID:   arg.VariantID,
			Quantity:    delta,
			Reason:      reason,
			Note:        arg.Note,
			CreatedBy:   arg.CreatedBy,
		})
		if err != nil {
			return err
		}
		result.Movement = movement

		after, err := q.GetInventoryStock(ctx, GetInventoryStockParams{
			InventoryID: arg.InventoryID,
			VariantID:   arg.VariantID,
		})
		if err != nil {
			return err
		}
		result.OnHand = after.Quantity
		return nil
	})

	return result, err
}

// nowNullTime is a small helper for the status-setting queries.
func nowNullTime() sql.NullTime {
	return sql.NullTime{Time: time.Now(), Valid: true}
}
