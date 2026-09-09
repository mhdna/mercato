package db

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/sqlc-dev/pqtype"
)

// recordAuditParams describes one create/update/delete on a non-inventory
// entity. Inventory stock changes use stock_movements instead (see
// tx_stock.go); this is for everything else -- clients, products,
// suppliers, expenses, price lists, settings.
type recordAuditParams struct {
	EntityType string
	EntityID   int64
	Action     AuditAction
	Before     any // nil on create
	After      any // nil on delete
	ActorID    sql.NullInt64
	Source     string // defaults to "ui"
}

// recordAudit writes one audit_logs row. Callers pass the transaction-bound
// *Queries so the audit row commits atomically with the change it records.
func (q *Queries) recordAudit(ctx context.Context, arg recordAuditParams) error {
	source := arg.Source
	if source == "" {
		source = "ui"
	}

	var before, after pqtype.NullRawMessage
	if arg.Before != nil {
		raw, err := json.Marshal(arg.Before)
		if err != nil {
			return err
		}
		before = pqtype.NullRawMessage{RawMessage: raw, Valid: true}
	}
	if arg.After != nil {
		raw, err := json.Marshal(arg.After)
		if err != nil {
			return err
		}
		after = pqtype.NullRawMessage{RawMessage: raw, Valid: true}
	}

	_, err := q.CreateAuditLog(ctx, CreateAuditLogParams{
		EntityType: arg.EntityType,
		EntityID:   arg.EntityID,
		Action:     arg.Action,
		Before:     before,
		After:      after,
		ActorID:    arg.ActorID,
		Source:     source,
	})
	return err
}
