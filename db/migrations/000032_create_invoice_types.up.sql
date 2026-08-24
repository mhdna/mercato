CREATE TABLE IF NOT EXISTS invoice_types (
    id bigserial primary key,
    name text not null,
    code text not null unique,
    is_default boolean not null default false,
    is_active boolean not null default true,
    created_at timestamp(0) with time zone not null default now()
);

INSERT INTO invoice_types (name, code, is_default, is_active)
VALUES ('Retail', 'RT', true, true);

ALTER TABLE invoices ADD COLUMN invoice_type_id bigint references invoice_types(id);
UPDATE invoices SET invoice_type_id = (SELECT id FROM invoice_types WHERE code = 'RT');
ALTER TABLE invoices ALTER COLUMN invoice_type_id SET NOT NULL;

ALTER TABLE invoice_indexes ADD COLUMN invoice_type_id bigint references invoice_types(id);
UPDATE invoice_indexes SET invoice_type_id = (SELECT id FROM invoice_types WHERE code = 'RT');
ALTER TABLE invoice_indexes ALTER COLUMN invoice_type_id SET NOT NULL;

ALTER TABLE invoice_indexes DROP CONSTRAINT invoice_indexes_pkey;
ALTER TABLE invoice_indexes ADD PRIMARY KEY (cashbox_id, year, invoice_type_id);
