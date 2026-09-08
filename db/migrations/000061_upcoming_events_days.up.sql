-- How many days ahead the appbar "Upcoming event" card scans the retail
-- calendar. Customisable from Settings > General.
ALTER TABLE app_settings
  ADD COLUMN upcoming_events_days SMALLINT NOT NULL DEFAULT 15;
