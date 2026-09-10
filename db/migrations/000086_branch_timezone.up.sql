-- The branch's IANA timezone (e.g. 'Asia/Beirut'). branch_invoices.occurred_at
-- is a UTC instant; the daily-income / sales-by-day rollups convert it to
-- this zone before taking ::date so a kashi day column matches the till's
-- own `date(created_at,'localtime')` view. 'UTC' keeps today's behaviour.
ALTER TABLE branches ADD COLUMN IF NOT EXISTS timezone text NOT NULL DEFAULT 'UTC';
