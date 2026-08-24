-- Brings kashi's cashbox_accounts up to the shape kashi-pos actually uses:
-- a payment method has a currency (currency_code), a display position
-- (sort_order, added in kashi-pos's 000067), and an optional accent color
-- (000068). kashi's version was previously just id+name. updated_at backs
-- the sync catch-up cursor (Phase 3).
ALTER TABLE cashbox_accounts ADD COLUMN currency_code text not null default 'USD' references currencies(code);
ALTER TABLE cashbox_accounts ADD COLUMN sort_order integer not null default 0;
ALTER TABLE cashbox_accounts ADD COLUMN color text not null default '';
ALTER TABLE cashbox_accounts ADD COLUMN updated_at timestamp(0) with time zone not null default now();

UPDATE cashbox_accounts SET sort_order = id;
