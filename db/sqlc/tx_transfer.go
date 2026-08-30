package db

import (
	"context"
	"database/sql"
	"errors"
)

// TransferItemParams is one line on a draft transfer.
type TransferItemParams struct {
	VariantID sql.NullInt64 `json:"variant_id"`
	AssetID   sql.NullInt64 `json:"asset_id"`
	Quantity  int64         `json:"quantity"`
}

type CreateTransferTxParams struct {
	FromInventoryID int64                `json:"from_inventory_id"`
	ToInventoryID   int64                `json:"to_inventory_id"`
	Type            TransferType         `json:"type"`
	Code            string               `json:"code"`
	Note            string               `json:"note"`
	Items           []TransferItemParams `json:"items"`
}

type CreateTransferTxResult struct {
	Transfer Transfer       `json:"transfer"`
	Items    []TransferItem `json:"items"`
}

// CreateTransferTx writes a draft transfer and its lines. No stock moves
// until the transfer is dispatched.
func (store *SQLStore) CreateTransferTx(ctx context.Context, arg CreateTransferTxParams) (CreateTransferTxResult, error) {
	var result CreateTransferTxResult

	if arg.FromInventoryID == arg.ToInventoryID {
		return result, errors.New("transfer source and destination must differ")
	}

	err := store.execTx(ctx, func(q *Queries) error {
		transfer, err := q.CreateTransfer(ctx, CreateTransferParams{
			FromInventoryID: arg.FromInventoryID,
			ToInventoryID:   arg.ToInventoryID,
			Type:            arg.Type,
			Code:            arg.Code,
			Note:            arg.Note,
		})
		if err != nil {
			return err
		}

		items := make([]TransferItem, 0, len(arg.Items))
		for _, it := range arg.Items {
			created, err := q.CreateTransferItem(ctx, CreateTransferItemParams{
				TransferID: transfer.ID,
				VariantID:  it.VariantID,
				AssetID:    it.AssetID,
				Quantity:   it.Quantity,
			})
			if err != nil {
				return err
			}
			items = append(items, created)
		}

		result.Transfer = transfer
		result.Items = items
		return nil
	})

	return result, err
}

type TransferStageTxParams struct {
	TransferID int64         `json:"transfer_id"`
	CreatedBy  sql.NullInt64 `json:"created_by"`
}

type TransferStageTxResult struct {
	Transfer  Transfer        `json:"transfer"`
	Movements []StockMovement `json:"movements"`
}

// TransferDispatchTx moves a draft transfer to 'dispatched': every product
// line leaves the source inventory as a 'transfer_out' movement. Stock in
// transit is simply absent from both locations until received.
func (store *SQLStore) TransferDispatchTx(ctx context.Context, arg TransferStageTxParams) (TransferStageTxResult, error) {
	var result TransferStageTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		transfer, err := q.GetTransfer(ctx, arg.TransferID)
		if err != nil {
			return err
		}
		if transfer.Status != TransferStatusDraft {
			return errors.New("only a draft transfer can be dispatched")
		}

		items, err := q.ListTransferItems(ctx, arg.TransferID)
		if err != nil {
			return err
		}

		for _, it := range items {
			if !it.VariantID.Valid {
				continue
			}
			movement, err := q.applyStockMovement(ctx, applyStockMovementParams{
				InventoryID:   transfer.FromInventoryID,
				VariantID:     it.VariantID.Int64,
				Quantity:      -it.Quantity,
				Reason:        StockMovementReasonTransferOut,
				ReferenceType: "transfer",
				ReferenceID:   transfer.ID,
				CreatedBy:     arg.CreatedBy,
			})
			if err != nil {
				return err
			}
			result.Movements = append(result.Movements, movement)
		}

		updated, err := q.SetTransferStatus(ctx, SetTransferStatusParams{
			Status:       TransferStatusDispatched,
			DispatchedAt: nowNullTime(),
			ID:           transfer.ID,
		})
		if err != nil {
			return err
		}
		result.Transfer = updated
		return nil
	})

	return result, err
}

// TransferReceiveTx moves a dispatched transfer to 'received': every
// product line lands in the destination inventory as a 'transfer_in'
// movement, carrying the source location's moving-average cost so the
// destination's average is blended correctly. The variant's global average
// is untouched -- a transfer moves neither total quantity nor total cost.
func (store *SQLStore) TransferReceiveTx(ctx context.Context, arg TransferStageTxParams) (TransferStageTxResult, error) {
	var result TransferStageTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		transfer, err := q.GetTransfer(ctx, arg.TransferID)
		if err != nil {
			return err
		}
		if transfer.Status != TransferStatusDispatched {
			return errors.New("only a dispatched transfer can be received")
		}

		items, err := q.ListTransferItems(ctx, arg.TransferID)
		if err != nil {
			return err
		}

		for _, it := range items {
			if !it.VariantID.Valid {
				continue
			}

			var unitCost sql.NullInt64
			src, err := q.GetInventoryStock(ctx, GetInventoryStockParams{
				InventoryID: transfer.FromInventoryID,
				VariantID:   it.VariantID.Int64,
			})
			if err == nil {
				unitCost = sql.NullInt64{Int64: src.AvgCost, Valid: true}
			} else if !errors.Is(err, sql.ErrNoRows) {
				return err
			}

			movement, err := q.applyStockMovement(ctx, applyStockMovementParams{
				InventoryID:   transfer.ToInventoryID,
				VariantID:     it.VariantID.Int64,
				Quantity:      it.Quantity,
				Reason:        StockMovementReasonTransferIn,
				ReferenceType: "transfer",
				ReferenceID:   transfer.ID,
				UnitCost:      unitCost,
				CreatedBy:     arg.CreatedBy,
			})
			if err != nil {
				return err
			}
			result.Movements = append(result.Movements, movement)
		}

		updated, err := q.SetTransferStatus(ctx, SetTransferStatusParams{
			Status:     TransferStatusReceived,
			ReceivedAt: nowNullTime(),
			ID:         transfer.ID,
		})
		if err != nil {
			return err
		}
		result.Transfer = updated
		return nil
	})

	return result, err
}
