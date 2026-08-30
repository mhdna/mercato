ALTER TABLE invoices
  ADD COLUMN salesperson_id BIGINT REFERENCES salespersons (id),
  ADD COLUMN loyalty_points_delta BIGINT NOT NULL DEFAULT 0;
