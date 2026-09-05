-- Widen branch_settings to the full set of per-branch config the admin can
-- now view/edit centrally: behaviour feature-flags, non-admin page-access
-- locks, printing, and the attendance device config. Every column here is
-- reported by the branch on each sync tick (branchSettingsPayload in
-- kashi-pos), and any of them can be pushed back down via an
-- update_settings branch command -- except the ones the branch is marked as
-- owning locally (managed_locally), which the admin UI then shows
-- read-only and omits from the command payload.
--
-- Device identifiers (printer_id / receipt_printer / screen_port) and the
-- attendance secrets are still branch-reported for display, but default to
-- "managed at the branch" so a single central value never stomps a
-- per-machine one by accident.

ALTER TABLE branch_settings
  -- behaviour feature flags (POS: feature_flags rows)
  ADD COLUMN custom_item_discounts_enabled     BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN custom_item_prices_enabled        BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN per_unit_item_prices_enabled      BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN price_change_manual_override_mode BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN invoice_keyboard_mode             BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN client_required                   BOOLEAN NOT NULL DEFAULT false,

  -- non-admin page-access locks (POS: page_unlock_* feature flags)
  ADD COLUMN page_unlock_clients      BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN page_unlock_inventory    BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN page_unlock_transfers    BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN page_unlock_attendance   BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN page_unlock_salespersons BOOLEAN NOT NULL DEFAULT false,

  -- printing (POS: settings columns) -- branch-wide, safe to centralize
  ADD COLUMN printer_size       TEXT             NOT NULL DEFAULT '',
  ADD COLUMN receipt_width      BIGINT           NOT NULL DEFAULT 0,
  ADD COLUMN receipt_height     BIGINT           NOT NULL DEFAULT 0,
  ADD COLUMN receipt_enabled    BOOLEAN          NOT NULL DEFAULT false,
  ADD COLUMN receipt_font       TEXT             NOT NULL DEFAULT '',
  ADD COLUMN receipt_body_font  TEXT             NOT NULL DEFAULT '',
  ADD COLUMN receipt_title_size DOUBLE PRECISION NOT NULL DEFAULT 0,
  ADD COLUMN receipt_body_size  DOUBLE PRECISION NOT NULL DEFAULT 0,
  ADD COLUMN receipt_cutoff     BOOLEAN          NOT NULL DEFAULT false,

  -- per-machine device identifiers -- reported for display, default local
  ADD COLUMN printer_id      TEXT NOT NULL DEFAULT '',
  ADD COLUMN receipt_printer TEXT NOT NULL DEFAULT '',
  ADD COLUMN screen_port     TEXT NOT NULL DEFAULT '',

  -- attendance device config (POS: attendance_config row + one flag)
  ADD COLUMN akuvox_ip                            TEXT    NOT NULL DEFAULT '',
  ADD COLUMN akuvox_username                      TEXT    NOT NULL DEFAULT '',
  ADD COLUMN attendance_enabled                   BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN attendance_duplicate_interval_seconds BIGINT NOT NULL DEFAULT 0,
  ADD COLUMN attendance_cashier_history           BOOLEAN NOT NULL DEFAULT false,

  -- set of group keys the branch owns locally; the admin UI renders those
  -- groups read-only and never includes them in an update_settings payload.
  -- Groups: 'secrets', 'device_ids' (extend as needed).
  ADD COLUMN managed_locally TEXT[] NOT NULL DEFAULT '{}';
