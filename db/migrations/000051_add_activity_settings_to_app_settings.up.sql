ALTER TABLE app_settings
  ADD COLUMN activity_message_seconds SMALLINT NOT NULL DEFAULT 8,
  ADD COLUMN activity_display_mode TEXT NOT NULL DEFAULT 'notification';
