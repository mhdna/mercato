# Centralizing branch settings + coupons — implementation plan

Spans three repos: **kashi** (Go central API), **kashi-ui** (Vue admin), **kashi-pos** (Go/Wails till).
Additive migrations only, each with a `.down.sql`.

## Status (2026-08-31)

- **Phase 1 (settings blob): done.** See the Phase 1 section below.
- **Admin/manager password from kashi: done.** kashi-pos `ForceSetAdminPassword` (force-sets
  the local `admin` user, no old password) + `admin_password` field on the update_settings
  command; manager password was already wired in Phase 1 (`attendance_approver_password` →
  `UpdateAttendanceApproverPassword`). Both are write-only fields in the branch-settings
  **Secrets** group. Plaintext over the command channel, hashed locally (kashi=bcrypt,
  kashi-pos=sha256, so a hash can't be shipped). Test: `TestExecuteUpdateSettingsCommand_WidenedFields`.
- **POS Users centralization: done** (central-authoritative roster + read-only till mirror).
  - central: `branch_users` table (000073), `api/branch_user.go` CRUD +
    `PUT /branch_users/:id/pin`, `branch_users` sync-down entity, `set_branch_user_pin`
    command type. PIN is never stored centrally.
  - kashi-pos: `users.remote_id` (000080), `applySyncedBranchUser` (match by remote_id, else
    adopt by username, else create with blank hash), `set_branch_user_pin` executor,
    `branch_user_updated` WS trigger. Test: `TestBranchUserRosterSyncAndPinCommand`.
  - kashi-ui: **Users** tab on `branch-settings.vue` + `useBranchUsers` composable.
  - kashi-pos settings.vue: synced users show a **kashi** chip, a PIN-only "Reset PIN"
    dialog, and no rename/role/delete (`ListUsers` now returns `isSynced`). The reset lives
    in the admin-only Security tab — its existing gate.
  - central tests: `api/branch_user_test.go` (list / create-enqueues-pin / bad-pin /
    bad-role / update / pin-enqueue-only / delete / not-found) +
    `TestPutBranchSettingsManagedLocally_OK`.
  - **Not done:** a *granular* `reset_user_pin` permission assignable to non-admin roles.
    kashi-pos has `permissions` / `role_permissions` tables but no permission-check code at
    all — building that framework is its own task. Reset is admin-gated today.
- **Coupons (Phase 2): being built in parallel by the user** — `000072_coupon_categories`,
  `api/coupon_category.go`, `useCouponCategories.js`, edits to `coupon.vue`/`useCoupons.js`.
  This plan's Phase 2 section is superseded by that work; not touched here.

---

## Background: how the branch⇄central pipe works today

One narrow contract, mirrored in three structs:

| Direction | File | Struct |
|---|---|---|
| branch → central (admin read) | `kashi/api/branch_settings.go` | `branchSettingsRequest` |
| branch → central (push every sync tick) | `kashi-pos/sync_settings.go` | `branchSettingsPayload` |
| central → branch (`update_settings` command) | `kashi-pos/commands.go` | `updateSettingsCommandPayload` |

- The branch **pushes its full settings snapshot** every tick (`PUT /branch/settings`).
- An admin edit is an **enqueued command** (`POST /branches/:id/commands`, type `update_settings`), not a direct write. The branch pulls pending commands (`GET /branch/commands/pending`), runs them through the *same local functions its own Save button calls*, and acks (`POST /branch/commands/:id/ack`).
- `executeUpdateSettingsCommand` reads the current row first and passes every field the payload doesn't carry through unchanged → **pointer / absent field = "leave as-is"**. This is the mechanism the "managed locally" toggle rides on.
- Sync-*down* of shared entities is `GET /branch/sync/changes?entity=X&since=T` (`kashi/api/branch_catchup.go`) — a `switch` over entity name. Today: currencies, cashbox_accounts, products, clients, branch_targets, salespersons, expense_categories, loan_categories.

Central `allowedBranchCommandTypes` currently: `update_settings`, `remote_return_invoice`.

---

## Cross-cutting: "managed centrally vs at the branch"

New column on central `branch_settings`: `managed_locally text[] not null default '{}'` (set of field/group keys the branch owns).

- New page shows a **Centrally / At the branch** switch per group. Groups: `secrets`, `device_ids`, and any others we choose to make opt-out-able.
- When a key is "at the branch": the page renders the branch-reported value **read-only** and **omits it** from the `update_settings` payload. No kashi-pos change needed for the toggle — the executor's absent-field semantics already do the right thing.
- The branch keeps pushing real values upward regardless, so the admin always sees current state.

---

## PHASE 1 — widen the settings blob

**STATUS: implemented (2026-08-31).** All three repos build; central `go test ./api`,
kashi-pos `go test ./...`, and `vite build` pass. Central migration `000071` applied to the
dev DB. New kashi-pos test: `TestExecuteUpdateSettingsCommand_WidenedFields`.

Deviations from the plan below:
- **POS admin login password: deferred.** kashi-pos's only setter is
  `ChangePassword(username, old, new)` which needs the current password — a remote force-set
  path doesn't exist yet. Not carried in the command payload, not shown in the UI.
- Secrets actually centralized: `akuvox_password`, `attendance_approver_password` only.
  They are write-only in the UI (never stored/echoed centrally); sent in the command payload
  only when the admin types a value and the `secrets` group is set to "Centrally".
- `managed_locally` groups implemented: `secrets`, `device_ids`. Set via
  `PUT /branches/:id/settings/managed_locally` (admin-only; not a branch command).

Scope: **Behavior**, **Security → Page access**, **Printing**, **Attendance**. (POS Users deferred. Coupons = Phase 2.)

### Fields

| Group | Fields | Central-eligible | Default source |
|---|---|---|---|
| Behavior (feature_flags 0/1) | `custom_item_discounts`, `custom_item_prices`, `per_unit_item_prices`, `price_change_manual_override_mode`, `invoice_keyboard_mode`, `client_required` (+ `product_search_enabled`, already wired) | yes | central |
| Page access (feature_flags) | `page_unlock_clients`, `page_unlock_inventory`, `page_unlock_transfers`, `page_unlock_attendance`, `page_unlock_salespersons` | yes | central |
| Printing (settings cols) | `printer_size`, `receipt_width`, `receipt_height`, `receipt_enabled`, `receipt_font`, `receipt_body_font`, `receipt_title_size`, `receipt_body_size`, `receipt_cutoff` | yes | central |
| Device IDs (settings cols) | `printer_id` / `receipt_printer`, `screen_port` | yes | **at the branch** |
| Attendance (`attendance_config` + flag) | `akuvox_ip`, `akuvox_username`, `attendance_enabled`, `duplicate_interval_seconds`, `attendance_cashier_history` | yes | central |
| Secrets | `akuvox_password`, `attendance_approver_password`, POS admin login password | yes | **at the branch** |

Secrets: pushed as plaintext over the command channel (already TLS + branch-auth), applied locally with the same hashing the local Save path uses. POS admin password needs a hash-scheme parity check between repos before enabling — if they differ, send plaintext and let kashi-pos hash.

### kashi (central)

1. Migration `NNN_branch_settings_full.up.sql`: add nullable columns to `branch_settings` for every new scalar/bool above + `managed_locally text[]`.
2. `api/branch_settings.go`: extend `branchSettingsRequest`, `putBranchSettings`, and the `UpsertBranchSettings` sqlc params/query to accept + persist them. `getBranchSettings` returns them.
3. sqlc regen (`make sqlc` / `sqlc generate`).
4. No change to `branch_command.go` — payload is opaque JSON built by the UI.

### kashi-ui

1. `src/pages/branch-settings.vue` (created this branch): add **Behavior**, **Printing**, **Attendance**, **Security → Page access** sections in the existing `field-grid` style.
2. Per-group **Centrally / At the branch** switch bound to `managed_locally`.
3. Payload builder for the `update_settings` command: include every field **except** those whose group is "at the branch"; send booleans as real booleans, pointers omitted when local.
4. `formFromSettings` extended for the new fields.

### kashi-pos

1. `sync_settings.go` `branchSettingsPayload` (push): report all new fields — flags via `GetFeatureFlag`, `attendance_config` columns, `settings` printing columns. `pushSettings` fills them from `db.Setting` / `attendance_config` / flag reads.
2. `commands.go` `updateSettingsCommandPayload` + `executeUpdateSettingsCommand`: add pointer fields; apply via existing local funcs —
   - behavior + `page_unlock_*` → `SetFeatureFlag`
   - `UpdateAttendanceConfig` (+ new setters for `duplicate_interval_seconds`, `approver_password`, and the `attendance_cashier_history` flag)
   - `UpdateSettings` / `UpdateReceiptSettings` for printing columns
   - each guarded `if p.X != nil`
3. No knowledge of `managed_locally` on this side — it just receives fewer fields.

### Tests
- kashi: `branch_settings` upsert/read round-trip for new columns.
- kashi-pos: `commands` test — an `update_settings` payload with each new field applies; an absent field leaves current value; a "local" field never arrives.

---

## PHASE 2 — coupons, modeled like price lists

Three levels: **coupon lists → coupon types → issued coupons**. Branch assignment at the **list** level (mirror `price_list_branches`). Coupon kinds: **one-time** (consumed on redemption) vs **reference** (stays valid, recorded per use, never consumed). All issuance and management is central; the till holds a read-only mirror and reports redemptions upward.

### Naming collision

The current central `coupons` table (`code, status, discount_type, reason, client_id, valid_until`, UI at `/pricing` → `coupons.vue`) is a **client discount voucher** feature, unused by invoice flows (`grep` in `sales_invoice.go` / `return_invoice.go` finds nothing). Plan: rename it to `client_vouchers` (table + `db/query/coupon.sql` + `api/coupon.go` + UI), free the `coupon*` namespace for the new model, and repoint `/pricing` at the new manager. **Open: confirm the old feature can be renamed/retired.**

### Central schema (`NNN_coupons.up.sql`)

```sql
create table coupon_lists (
  id bigserial primary key,
  name text not null,
  is_active boolean not null default true,
  valid_from timestamptz not null default now(),
  valid_to   timestamptz not null default now(),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table coupon_list_branches (          -- one list per branch, like price_list_branches
  branch_id      bigint primary key references branches(id) on delete cascade,
  coupon_list_id bigint not null references coupon_lists(id) on delete cascade,
  created_at     timestamptz not null default now()
);
create index on coupon_list_branches (coupon_list_id);

create table coupon_types (
  id             bigserial primary key,
  coupon_list_id bigint not null references coupon_lists(id) on delete cascade,
  name           text not null,
  value_cents    bigint not null check (value_cents > 0),
  currency_code  text not null default 'USD',
  kind           text not null check (kind in ('one_time','reference')),
  is_active      boolean not null default true,
  created_at     timestamptz not null default now(),
  updated_at     timestamptz not null default now(),
  unique (coupon_list_id, name)
);

create table coupons (                        -- issued codes
  code            text primary key,
  coupon_type_id  bigint not null references coupon_types(id) on delete cascade,
  value_cents     bigint not null,            -- snapshot from type at issue time
  is_active       boolean not null default true,
  redeemed_invoice_ref text,                  -- branch client_ref that consumed it (one_time); null = open
  redeemed_branch_id   bigint references branches(id),
  redeemed_at          timestamptz,
  created_at      timestamptz not null default now(),
  updated_at      timestamptz not null default now()
);
create index on coupons (coupon_type_id);
```

`updated_at` on `coupon_types` / `coupons` / `coupon_lists` is the sync-down cursor. Add triggers or set it in every mutating query.

### Central API (mirror price-list routes)

```
POST/GET/PUT/DELETE  /coupon_lists          + GET /coupon_lists/:id
GET/PUT              /coupon_lists/:id/branches         (reuse BranchAssignDialog contract)
POST                /coupon_types                       (body carries coupon_list_id)
GET                 /coupon_lists/:id/types
PUT                 /coupon_types
DELETE              /coupon_types/:id       + POST /coupon_types/:id/bulk_delete
POST                /coupon_types/:id/issue             ({count} | {codes:[...]}) → creates coupons
GET                 /coupon_types/:id/coupons           (ServerSideTable paged, root_key "coupons")
PUT                 /coupons/:code                      (toggle is_active)
DELETE              /coupons/:code
```

Only one `coupon_list` per branch is enforced by the `coupon_list_branches` PK, same as price lists.

### Sync-down

`api/branch_catchup.go` switch — three new entities scoped to the caller branch's assigned list:

- `coupon_lists` → the one row assigned to this branch (via `coupon_list_branches`), if `is_active`
- `coupon_types` → types of that list updated since cursor
- `coupons` → issued coupons of those types updated since cursor

### Redemption path (branch → central)

At checkout the till validates the entered code against its **local mirror** (`is_active`, kind, and for `one_time` that `redeemed_*` is unset locally), records `invoice_coupon_payments` locally as it does now, and reports it up. Two options — pick during impl:

- **A (piggyback):** add `coupon_code` (+ amount) to `branchInvoiceRequest.payments` / a new field; central marks a `one_time` coupon redeemed when it ingests that branch invoice.
- **B (dedicated):** `POST /branch/coupon_redemptions {code, invoice_ref, amount, occurred_at}`.

B is cleaner (redemption isn't always tied to a sale we ingest, and reversY on return is explicit). Central sets `redeemed_*`; on a later sync the till sees the coupon flip to consumed.

**Double-spend window:** a list assigned to multiple branches means a `one_time` code physically exists at >1 till and could be redeemed offline at each before either syncs. Mitigations (decide): (a) warn in the UI that `one_time` types only make sense on single-branch lists; (b) central rejects the 2nd redemption on ingest and pushes a "coupon already consumed" state the till surfaces as an after-the-fact correction. Start with (a) + (b)-reject-and-log.

### kashi-pos

1. Migrations mirroring `coupon_lists` / `coupon_types` / `coupons` as **read-only mirror** tables (like the currencies mirror).
2. Sync worker: pull the three entities each tick; upsert by pk; respect `updated_at` cursor.
3. Checkout: resolve codes from the mirror; keep local `invoice_coupon_payments`; emit redemption via chosen path.
4. Return/exchange: reversing a `one_time` redemption emits a redemption-reversal (central clears `redeemed_*`).
5. Settings **Coupons** tab: remove the local `coupon_types` editor and the issue-coupon UI; optionally keep a read-only list of what's synced.
6. Retire local-only coupon creation queries.

### kashi-ui

1. New `CouponListManager.vue` closely following `PriceDiscountListManager.vue`: left sidebar = coupon lists; detail = **types** table for the selected list (name, value, kind, active) with add/edit/delete; opening a type drills into its **issued coupons** (ServerSideTable: code, value, active, redeemed-at/where) with an **Issue** action (choose kind is fixed by the type; enter count or explicit codes) and per-row enable/disable.
2. Reuse `BranchAssignDialog` for list→branches.
3. `/pricing` route → `CouponListManager`; move old client-voucher screen to its own route if kept.
4. Composables mirroring `usePriceLists` / `useValueListStore`.

### Tests
- kashi: coupon list/type/coupon CRUD; issue N; one-time redemption sets `redeemed_*`; 2nd redemption rejected; sync-down scoping returns only the branch's assigned list.
- kashi-pos: mirror upsert; checkout resolves a synced coupon; redemption emitted; reversal on return.

---

## Sequencing

1. **Phase 1** end-to-end (all three repos) + tests. Ship.
2. **Phase 2a**: central schema + API + `client_vouchers` rename + UI manager (no branch sync yet — central-only).
3. **Phase 2b**: sync-down entities + kashi-pos mirror + read path.
4. **Phase 2c**: redemption reporting + reversal + double-spend handling.
5. Later: POS Users centralization (deferred).
