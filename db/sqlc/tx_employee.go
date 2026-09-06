package db

import "context"

// BatchSetEmployeeSalaries applies a batch of payroll edits from the
// admin "edit salaries" table in one transaction, so a partial save can
// never leave half the roster on new figures and half on old ones -- if
// any row's id is unknown (0 rows updated) or violates a check constraint,
// the whole batch rolls back. Returns the updated rows in input order.
func (store *SQLStore) BatchSetEmployeeSalaries(ctx context.Context, args []UpdateEmployeeSalaryParams) ([]Employee, error) {
	updated := make([]Employee, 0, len(args))

	err := store.execTx(ctx, func(q *Queries) error {
		for _, arg := range args {
			employee, err := q.UpdateEmployeeSalary(ctx, arg)
			if err != nil {
				return err
			}
			updated = append(updated, employee)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}
