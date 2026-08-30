CREATE TABLE app_settings (
  id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
  financials_high_season_months JSONB NOT NULL DEFAULT '[2, 5, 6, 9, 11, 12]',
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO app_settings (id) VALUES (1);
