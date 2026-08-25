alter table expenses add column category text not null default '';
update expenses set category = ec.name
from expense_categories ec
where ec.id = expenses.category_id;
alter table expenses drop column category_id;

alter table branch_expenses add column category text not null default '';
update branch_expenses set category = ec.name
from expense_categories ec
where ec.id = branch_expenses.category_id;
alter table branch_expenses drop column category_id;

alter table recurring_expenses add column category text not null default '';
update recurring_expenses set category = ec.name
from expense_categories ec
where ec.id = recurring_expenses.category_id;
alter table recurring_expenses drop column category_id;
