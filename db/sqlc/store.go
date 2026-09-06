package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store interface {
	Querier
	SalesInvoiceTx(ctx context.Context, arg SalesInvoiceTxParams) (SalesInvoiceTxResult, error)
	CreateProductTx(ctx context.Context, arg CreateProductTxParams) (CreateProductTxResult, error)
	UpdateProductTx(ctx context.Context, arg UpdateProductTxParams) (UpdateProductTxResult, error)
	AddVariantTx(ctx context.Context, arg AddVariantTxParams) (ProductVariant, error)
	ReturnInvoiceTx(ctx context.Context, arg ReturnInvoiceTxParams) (ReturnInvoiceTxResult, error)
	CreateBranchTx(ctx context.Context, arg CreateBranchTxParams) (CreateBranchTxResult, error)
	CreatePurchaseTx(ctx context.Context, arg CreatePurchaseTxParams) (CreatePurchaseTxResult, error)
	PurchaseReceiveTx(ctx context.Context, arg PurchaseReceiveTxParams) (PurchaseReceiveTxResult, error)
	CreateTransferTx(ctx context.Context, arg CreateTransferTxParams) (CreateTransferTxResult, error)
	TransferDispatchTx(ctx context.Context, arg TransferStageTxParams) (TransferStageTxResult, error)
	TransferReceiveTx(ctx context.Context, arg TransferStageTxParams) (TransferStageTxResult, error)
	StockAdjustmentTx(ctx context.Context, arg StockAdjustmentTxParams) (StockAdjustmentTxResult, error)
	CreateBranchInvoiceTx(ctx context.Context, arg CreateBranchInvoiceTxParams) (CreateBranchInvoiceTxResult, error)
	ApplyBranchSettlementTx(ctx context.Context, arg ApplyBranchSettlementTxParams) (ApplyBranchSettlementTxResult, error)
	FireRecurringExpenseTx(ctx context.Context, recurring RecurringExpense) (Expense, error)
	DeleteBranchTargetSeriesTx(ctx context.Context, seriesID int64) (BranchTargetSeries, error)
	BatchSetEmployeeSalaries(ctx context.Context, args []UpdateEmployeeSalaryParams) ([]Employee, error)
}

// provides all the functions to execute SQL queries
type SQLStore struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

// type TransferTxParams struct {
// 	FromInventoryID int64          `json:"from_inventory_id"`
// 	ToInventoryID   int64          `json:"to_inventory_id"`
// 	Items           []TransferItem `json:"items"`
// 	TransferType    TransferType   `json:"type"`
// }

// type TransferTxResult struct {
// 	Transfer      Transfer       `json:"transfer"`
// 	FromInventory Inventory      `json:"from_inventory"`
// 	ToInventory   Inventory      `json:"to_inventory"`
// 	Items         []TransferItem `json:"items"`
// }

// // TransferTx performs an inventory transfer (assets/products) from one inventory to another.
// // It creates a transfer record, adds inventory entries, and updates inventory balance within a single db transaction.
// func (store *SQLStore) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
// 	var result TransferTxResult

// 	err := store.execTx(ctx, func(q *Queries) error {
// 		var err error

// 		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
// 			FromInventoryID: arg.FromInventoryID,
// 			ToInventoryID:   arg.ToInventoryID,
// 			Type:            arg.TransferType,
// 		})
// 		if err != nil {
// 			return err
// 		}

// 		for _, i := range arg.Items {
// 			var productID, assetID sql.NullInt64
// 			switch arg.TransferType {
// 			case TransferTypeAssets:
// 				assetID = i.AssetID
// 			case TransferTypeProducts:
// 				productID = i.ProductID
// 			default:
// 				return fmt.Errorf("unknown transfer type: %s", arg.TransferType)
// 			}

// 			_, err := q.CreateTransferItem(ctx, CreateTransferItemParams{
// 				TransferID: result.Transfer.ID,
// 				ProductID:  productID,
// 				AssetID:    assetID,
// 				Quantity:   i.Quantity,
// 			})
// 			if err != nil {
// 				return err
// 			}
// 			_, err = q.CreateEntryItem(ctx, CreateEntryItemParams{
// 				InventoryID:   arg.FromInventoryID,
// 				ReferenceType: EntryReferenceTypeTransfer,
// 				ReferenceID:   result.Transfer.ID,
// 				ProductID:     productID,
// 				AssetID:       assetID,
// 				Quantity:      i.Quantity,
// 			})
// 			if err != nil {
// 				return err
// 			}
// 			_, err = q.CreateEntryItem(ctx, CreateEntryItemParams{
// 				InventoryID:   arg.FromInventoryID,
// 				ReferenceType: EntryReferenceTypeTransfer,
// 				ReferenceID:   result.Transfer.ID,
// 				ProductID:     productID,
// 				AssetID:       assetID,
// 				Quantity:      -i.Quantity,
// 			})
// 			if err != nil {
// 				return err
// 			}

// 			// TODO: update inventories
// 		}
// 		return nil
// 	})

// 	return result, err
// }
