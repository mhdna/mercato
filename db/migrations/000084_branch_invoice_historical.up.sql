-- Marks an invoice that a branch pushed as a one-time historical backfill
-- (a sale rung up before central sync was switched on). Such invoices are
-- recorded in full -- header, items, payments, loyalty -- so admin reports
-- and totals are complete, but by default they do NOT move central stock:
-- kashi's inventory was almost certainly seeded at a point well after the
-- oldest backfilled sale, so replaying years of decrements would drain it.
-- The branch's "sync everything" flow lets the operator opt back into stock
-- movement per run when the timeline actually lines up.
ALTER TABLE branch_invoices
    ADD COLUMN IF NOT EXISTS historical boolean NOT NULL DEFAULT false;
