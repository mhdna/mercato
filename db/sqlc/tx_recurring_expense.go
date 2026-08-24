package db

import (
	"context"
	"database/sql"
	"time"
)

// FireRecurringExpenseTx creates the real expense row for a due recurring
// template and advances its next_due_at in the same transaction, so a
// crash between the two can never leave a template stuck re-firing the
// same due date on the next scheduler tick (which would create duplicate
// expenses) or advanced with no expense created (which would silently
// drop one).
func (store *SQLStore) FireRecurringExpenseTx(ctx context.Context, recurring RecurringExpense) (Expense, error) {
	var expense Expense

	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		expense, err = q.CreateExpense(ctx, CreateExpenseParams{
			Description:        recurring.Description,
			Category:           recurring.Category,
			Amount:             recurring.Amount,
			CurrencyCode:       recurring.CurrencyCode,
			RecurringExpenseID: sql.NullInt64{Int64: recurring.ID, Valid: true},
		})
		if err != nil {
			return err
		}

		_, err = q.UpdateRecurringExpenseNextDue(ctx, UpdateRecurringExpenseNextDueParams{
			ID:        recurring.ID,
			NextDueAt: nextDueAfter(recurring.NextDueAt, recurring.IntervalUnit, recurring.IntervalCount),
		})
		return err
	})

	return expense, err
}

// nextDueAfter advances a due date by the template's interval. Calendar
// month arithmetic (not a fixed 30-day approximation) so "every 5 months"
// lands on the same day-of-month each time, the way a human reading
// "every 5 months" expects.
func nextDueAfter(from time.Time, unit string, count int32) time.Time {
	if unit == "month" {
		return from.AddDate(0, int(count), 0)
	}
	return from.AddDate(0, 0, int(count))
}
