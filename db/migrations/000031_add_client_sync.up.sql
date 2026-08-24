ALTER TABLE clients ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

CREATE TABLE client_links (
  branch_id BIGINT NOT NULL REFERENCES branches (id),
  branch_client_id BIGINT NOT NULL,
  client_id BIGINT NOT NULL REFERENCES clients (id),
  PRIMARY KEY (branch_id, branch_client_id)
);

CREATE INDEX idx_client_links_client_id ON client_links (client_id);
