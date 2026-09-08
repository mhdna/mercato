-- name: ListCalendarEvents :many
SELECT * FROM calendar_events
ORDER BY start_date, id;

-- name: CreateCalendarEvent :one
INSERT INTO calendar_events (
  name,
  icon,
  color,
  start_date,
  end_date
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: DeleteCalendarEvent :exec
DELETE FROM calendar_events
WHERE id = $1;
