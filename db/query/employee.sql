-- name: CreateEmployee :one
INSERT INTO employees (
  branch_id,
  name,
  role,
  status,
  hired_on,
  base_salary_cents,
  commission_cents,
  late_deduction_rate_cents,
  late_deduction_unit,
  late_units_centi
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: GetEmployee :one
SELECT * FROM employees
WHERE id = $1 LIMIT 1;

-- name: UpdateEmployee :one
UPDATE employees
SET branch_id = $2,
    name = $3,
    role = $4,
    status = $5,
    hired_on = $6,
    base_salary_cents = $7,
    commission_cents = $8,
    late_deduction_rate_cents = $9,
    late_deduction_unit = $10,
    late_units_centi = $11,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateEmployeeSalary :one
UPDATE employees
SET base_salary_cents = $2,
    commission_cents = $3,
    late_deduction_rate_cents = $4,
    late_deduction_unit = $5,
    late_units_centi = $6,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteEmployee :exec
DELETE FROM employees WHERE id = $1;

-- name: ListEmployees :many
SELECT * FROM employees
WHERE (
    BTRIM(sqlc.arg(search)::text) = ''
    OR name ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%'
    OR role ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%'
  )
  AND (sqlc.narg(branch_id)::bigint IS NULL OR branch_id = sqlc.narg(branch_id))
ORDER BY name
LIMIT sqlc.arg(page_size)
OFFSET sqlc.arg(page_offset);

-- name: CountEmployees :one
SELECT COUNT(*) FROM employees
WHERE (
    BTRIM(sqlc.arg(search)::text) = ''
    OR name ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%'
    OR role ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%'
  )
  AND (sqlc.narg(branch_id)::bigint IS NULL OR branch_id = sqlc.narg(branch_id));
