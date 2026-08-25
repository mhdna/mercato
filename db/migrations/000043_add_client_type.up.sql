-- Wholesale clients are a central-office/B2B concept -- they buy through
-- kashi directly, never at a branch till, so this column also drives what
-- ListClientsUpdatedSince exposes to branches (retail only; see
-- api/branch_catchup.go and kashi-pos's client sync, which has no concept
-- of wholesale at all).
ALTER TABLE clients ADD COLUMN client_type text NOT NULL DEFAULT 'retail'
    CHECK (client_type IN ('retail', 'wholesale'));

CREATE INDEX ON clients (client_type);
