package db

import (
	"context"
	"database/sql"
)

// DeleteBranchTargetSeriesTx removes a recurring series and every period it
// generated. Deactivating the occurrences before deleting the series row
// matters: branch_targets.series_id is ON DELETE SET NULL, so deleting the
// series first would silently orphan those rows as still-active fixed
// targets instead of removing them.
func (store *SQLStore) DeleteBranchTargetSeriesTx(ctx context.Context, seriesID int64) (BranchTargetSeries, error) {
	var series BranchTargetSeries

	err := store.execTx(ctx, func(q *Queries) error {
		if err := q.DeactivateBranchTargetsBySeries(ctx, sql.NullInt64{Int64: seriesID, Valid: true}); err != nil {
			return err
		}

		var err error
		series, err = q.DeleteBranchTargetSeries(ctx, seriesID)
		return err
	})

	return series, err
}
