ALTER TABLE app_settings
  DROP COLUMN upcoming_events_menu_days,
  DROP COLUMN hidden_builtin_events;

DROP TABLE calendar_events;
