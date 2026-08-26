-- Receipt photos attached to a branch expense after the fact, via the
-- one-time QR upload flow (api/expense_upload.go). branch_expense_id is a
-- real foreign key here (unlike branch_expenses' own opaque branch-local
-- ids) because both tables live in kashi's own schema.
create table if not exists branch_expense_images (
    id bigserial primary key,
    branch_expense_id bigint not null references branch_expenses(id),
    file_path text not null,
    content_type text not null,
    size_bytes bigint not null,
    created_at timestamp(0) with time zone not null default now()
);
create index if not exists idx_branch_expense_images_branch_expense_id on branch_expense_images (branch_expense_id);

-- One-time, expiring tokens minted alongside a branch expense (see
-- createBranchExpense in api/branch_expense.go) so a phone can be handed a
-- link -- via QR code shown in kashi-pos -- to upload receipt photos
-- without any login. used_at enforces single-use: once set, the token is
-- permanently dead rather than reusable until it expires.
create table if not exists branch_expense_upload_tokens (
    id bigserial primary key,
    branch_expense_id bigint not null references branch_expenses(id),
    token text not null unique,
    expires_at timestamp(0) with time zone not null,
    used_at timestamp(0) with time zone,
    created_at timestamp(0) with time zone not null default now()
);
