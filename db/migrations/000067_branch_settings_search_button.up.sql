-- Whether the branch's POS shows the product-search button on its sale
-- screen. Unlike every other column here (which the branch owns and only
-- pushes up), this one is also admin-editable from kashi via an
-- update_settings branch command. Defaults to true, matching POS behavior
-- before the toggle existed.
ALTER TABLE branch_settings
  ADD COLUMN search_button_enabled BOOLEAN NOT NULL DEFAULT true;
