-- Populate expense_categories from whatever distinct, non-blank category
-- text already exists across all three expense tables, then point each
-- table at it by id and drop the old free-text column. category_id stays
-- nullable (unlike loans.category_id) -- a blank category was always a
-- valid "uncategorized" state here, and this migration preserves that
-- rather than forcing every historical row into a category it never had.
insert into expense_categories (name)
select distinct category from expenses where category <> ''
union
select distinct category from branch_expenses where category <> ''
union
select distinct category from recurring_expenses where category <> ''
on conflict (name) do nothing;

alter table expenses add column category_id bigint references expense_categories(id);
update expenses set category_id = ec.id
from expense_categories ec
where ec.name = expenses.category and expenses.category <> '';
alter table expenses drop column category;

alter table branch_expenses add column category_id bigint references expense_categories(id);
update branch_expenses set category_id = ec.id
from expense_categories ec
where ec.name = branch_expenses.category and branch_expenses.category <> '';
alter table branch_expenses drop column category;

alter table recurring_expenses add column category_id bigint references expense_categories(id);
update recurring_expenses set category_id = ec.id
from expense_categories ec
where ec.name = recurring_expenses.category and recurring_expenses.category <> '';
alter table recurring_expenses drop column category;
