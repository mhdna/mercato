-- name: CreateShift :one
INSERT INTO shifts (
  cashbox_id
) 
VALUES ( $1 )
RETURNING *;

-- name: GetShift :one
SELECT * FROM shifts
WHERE id = $1
LIMIT 1;

-- name: CloseShift :exec
UPDATE shifts 
  SET closed_at = NOW(),
  is_closed = true
WHERE id = $1;


-- name: ListShifts :many
SELECT * FROM shifts
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: CountShifts :one
SELECT COUNT(*) FROM shifts;
