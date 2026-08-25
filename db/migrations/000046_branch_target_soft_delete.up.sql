-- branch_targets was the one synced entity still doing a hard DELETE --
-- every other synced entity (currencies, cashbox_accounts, salespersons)
-- uses an is_active flag instead, because the sync mechanism is purely
-- "list rows changed since X": a hard-deleted row just vanishes from that
-- query, so a branch never learns to remove its local copy. Bringing
-- targets in line with that convention (see api/branch_target.go's
-- deleteBranchTarget, updated to flip this instead of deleting).
alter table branch_targets add column is_active boolean not null default true;
