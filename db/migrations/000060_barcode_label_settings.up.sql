-- Default barcode-label page size (in PDF points, 1pt = 1/72in) used by the
-- barcode print endpoint. Overridable per print request.
ALTER TABLE app_settings
  ADD COLUMN barcode_label_width  REAL NOT NULL DEFAULT 288,
  ADD COLUMN barcode_label_height REAL NOT NULL DEFAULT 144;
