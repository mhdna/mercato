package db

import (
	"context"
	"database/sql"
	"errors"
)

// StockCountLineParams is one line on a draft stock count sheet: the SKU
// and what was physically counted. system_quantity is snapshotted from
// live on-hand at save time, not supplied by the caller.
type StockCountLineParams struct {
	VariantID       int64 `json:"variant_id"`
	CountedQuantity int64 `json:"counted_quantity"`
}

type SaveStockCountTxParams struct {
	ID          int64 // 0 => create a new draft
	InventoryID int64
	Note        string
	CreatedBy   sql.NullInt64
	Items       []StockCountLineParams
}

// SaveStockCountTx writes (or rewrites) a draft stock count's header and
// lines. It re-snapshots system_quantity from live on-hand and recomputes
// difference every time -- saving never touches stock. Editing a count
// that is no longer a draft is refused.
func (store *SQLStore) SaveStockCountTx(ctx context.Context, arg SaveStockCountTxParams) (StockCount, error) {
	var result StockCount

	err := store.execTx(ctx, func(q *Queries) error {
		var count StockCount
		var err error
		if arg.ID == 0 {
			count, err = q.CreateStockCount(ctx, CreateStockCountParams{
				InventoryID: arg.InventoryID,
				Note:        arg.Note,
				CreatedBy:   arg.CreatedBy,
			})
			if err != nil {
				return err
			}
		} else {
			count, err = q.GetStockCount(ctx, arg.ID)
			if err != nil {
				return err
			}
			if count.Status != StockCountStatusDraft {
				return errors.New("only a draft stock count can be edited")
			}
			if err := q.SetStockCountNote(ctx, SetStockCountNoteParams{ID: count.ID, Note: arg.Note}); err != nil {
				return err
			}
			if err := q.DeleteStockCountItems(ctx, count.ID); err != nil {
				return err
			}
		}

		for _, item := range arg.Items {
			stock, err := q.GetInventoryStock(ctx, GetInventoryStockParams{
				InventoryID: arg.InventoryID,
				VariantID:   item.VariantID,
			})
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return err
			}
			systemQty := stock.Quantity // zero value when no row exists yet

			if _, err := q.AddStockCountItem(ctx, AddStockCountItemParams{
				StockCountID:    count.ID,
				VariantID:       item.VariantID,
				SystemQuantity:  systemQty,
				CountedQuantity: item.CountedQuantity,
				Difference:      item.CountedQuantity - systemQty,
			}); err != nil {
				return err
			}
		}

		result = count
		return nil
	})

	return result, err
}

type PostStockCountTxResult struct {
	StockCount StockCount      `json:"stock_count"`
	Movements  []StockMovement `json:"movements"`
}

// PostStockCountTx moves a draft count to 'posted'. For each line it
// compares the counted quantity against *live* on-hand (not the snapshot
// taken at save time, which may be stale by now) and writes exactly one
// applyStockMovement covering the difference -- a line whose live on-hand
// already matches the count writes no movement. Metadata records both
// figures so drift between counting and posting stays visible.
func (store *SQLStore) PostStockCountTx(ctx context.Context, id int64, postedBy sql.NullInt64) (PostStockCountTxResult, error) {
	var result PostStockCountTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		count, err := q.GetStockCount(ctx, id)
		if err != nil {
			return err
		}
		if count.Status != StockCountStatusDraft {
			return errors.New("only a draft stock count can be posted")
		}

		items, err := q.ListStockCountItems(ctx, id)
		if err != nil {
			return err
		}

		for _, item := range items {
			liveStock, err := q.GetInventoryStock(ctx, GetInventoryStockParams{
				InventoryID: count.InventoryID,
				VariantID:   item.VariantID,
			})
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return err
			}
			liveOnHand := liveStock.Quantity
			delta := item.CountedQuantity - liveOnHand
			if delta == 0 {
				continue
			}

			movement, err := q.applyStockMovement(ctx, applyStockMovementParams{
				InventoryID:   count.InventoryID,
				VariantID:     item.VariantID,
				Quantity:      delta,
				Reason:        StockMovementReasonCount,
				ReferenceType: "stock_count",
				ReferenceID:   count.ID,
				CreatedBy:     postedBy,
				Metadata: map[string]any{
					"saved_system_quantity": item.SystemQuantity,
					"counted_quantity":      item.CountedQuantity,
				},
			})
			if err != nil {
				return err
			}
			result.Movements = append(result.Movements, movement)
		}

		posted, err := q.SetStockCountStatus(ctx, SetStockCountStatusParams{
			ID:       count.ID,
			Status:   StockCountStatusPosted,
			PostedBy: postedBy,
		})
		if err != nil {
			return err
		}
		result.StockCount = posted
		return nil
	})

	return result, err
}

// CancelStockCountTx discards a draft count. A count that has already been
// posted has moved stock and is not reversible through this path.
func (store *SQLStore) CancelStockCountTx(ctx context.Context, id int64) (StockCount, error) {
	var result StockCount

	err := store.execTx(ctx, func(q *Queries) error {
		count, err := q.GetStockCount(ctx, id)
		if err != nil {
			return err
		}
		if count.Status != StockCountStatusDraft {
			return errors.New("only a draft stock count can be cancelled")
		}

		cancelled, err := q.SetStockCountStatus(ctx, SetStockCountStatusParams{
			ID:     count.ID,
			Status: StockCountStatusCancelled,
		})
		if err != nil {
			return err
		}
		result = cancelled
		return nil
	})

	return result, err
}
