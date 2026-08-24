ALTER TABLE invoice_indexes DROP CONSTRAINT invoice_indexes_pkey;
ALTER TABLE invoice_indexes ADD PRIMARY KEY (cashbox_id, year);
ALTER TABLE invoice_indexes DROP COLUMN invoice_type_id;

ALTER TABLE invoices DROP COLUMN invoice_type_id;

DROP TABLE invoice_types;
