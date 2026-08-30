package db

import (
	"context"
	"fmt"
)

type CreateBranchTxParams struct {
	Name       string `json:"name"`
	Code       string `json:"code"`
	APIKeyHash string `json:"api_key_hash"`
}

type CreateBranchTxResult struct {
	Branch    Branch    `json:"branch"`
	Inventory Inventory `json:"inventory"`
}

// CreateBranchTx creates a branch together with its own store inventory
// and links the two, so branch (POS) sales can decrement central stock for
// that location from day one.
func (store *SQLStore) CreateBranchTx(ctx context.Context, arg CreateBranchTxParams) (CreateBranchTxResult, error) {
	var result CreateBranchTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		branch, err := q.CreateBranch(ctx, CreateBranchParams{
			Name:       arg.Name,
			Code:       arg.Code,
			ApiKeyHash: arg.APIKeyHash,
		})
		if err != nil {
			return err
		}

		inventory, err := q.CreateInventory(ctx, CreateInventoryParams{
			Name: fmt.Sprintf("Branch: %s", arg.Name),
			Type: InventoryTypeStore,
			Code: fmt.Sprintf("BRANCH-%s", arg.Code),
		})
		if err != nil {
			return err
		}

		if err := q.SetBranchInventory(ctx, SetBranchInventoryParams{
			ID:          branch.ID,
			InventoryID: nullInt64(inventory.ID),
		}); err != nil {
			return err
		}
		branch.InventoryID = nullInt64(inventory.ID)

		result.Branch = branch
		result.Inventory = inventory
		return nil
	})

	return result, err
}
