-- User-defined one-off calendar events, merged at render time with the
-- built-in retail calendar (ui/src/data/retailCalendarEvents.js).
CREATE TABLE calendar_events (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  icon TEXT NOT NULL DEFAULT 'mdi-calendar-star',
  color TEXT NOT NULL DEFAULT 'primary',
  start_date DATE NOT NULL,
  end_date DATE NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT calendar_events_dates_ordered CHECK (end_date >= start_date)
);

-- hidden_builtin_events: names of built-in retail events the user has
-- chosen to hide. upcoming_events_menu_days: how far ahead the appbar
-- "Upcoming events" dropdown scans (the inline pill uses the shorter
-- upcoming_events_days).
ALTER TABLE app_settings
  ADD COLUMN upcoming_events_menu_days SMALLINT NOT NULL DEFAULT 180,
  ADD COLUMN hidden_builtin_events JSONB NOT NULL DEFAULT '[]';
