CREATE TABLE branch_settings (
  branch_id BIGINT PRIMARY KEY REFERENCES branches (id),
  branch_name TEXT NOT NULL DEFAULT '',
  tax_rate DOUBLE PRECISION NOT NULL DEFAULT 0,
  rounding_mode TEXT NOT NULL DEFAULT 'half_up',
  rounding_currency TEXT NOT NULL DEFAULT '',
  exchange_rate BIGINT NOT NULL DEFAULT 0,
  exchange_window_hours BIGINT NOT NULL DEFAULT 0,
  market_name TEXT NOT NULL DEFAULT '',
  market_phone TEXT NOT NULL DEFAULT '',
  market_description TEXT NOT NULL DEFAULT '',
  return_policy TEXT NOT NULL DEFAULT '',
  website TEXT NOT NULL DEFAULT '',
  instagram TEXT NOT NULL DEFAULT '',
  social_platforms JSONB NOT NULL DEFAULT '[]',
  social_handles JSONB NOT NULL DEFAULT '{}',
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
