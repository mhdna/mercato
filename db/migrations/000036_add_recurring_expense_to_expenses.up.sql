-- Links an auto-fired expense back to the recurring template that created
-- it. NULL for every expense created manually (the existing path).
ALTER TABLE expenses ADD COLUMN recurring_expense_id bigint references recurring_expenses(id);
