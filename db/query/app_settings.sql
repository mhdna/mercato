-- name: GetAppSettings :one
SELECT * FROM app_settings WHERE id = 1;

-- name: UpdateAppSettings :one
UPDATE app_settings
SET financials_high_season_months = $1,
    activity_message_seconds = $2,
    activity_display_mode = $3,
    barcode_label_width = $4,
    barcode_label_height = $5,
    updated_at = now()
WHERE id = 1
RETURNING *;
