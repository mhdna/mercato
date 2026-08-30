-- name: GetDiscountListBranch :one
SELECT * FROM discount_list_branches
WHERE branch_id = $1 LIMIT 1;

-- name: ListDiscountListBranchIDs :many
SELECT branch_id FROM discount_list_branches
WHERE discount_list_id = $1
ORDER BY branch_id;

-- name: ListBranchesForDiscountList :many
SELECT b.id, b.name, b.code
FROM discount_list_branches dlb
JOIN branches b ON b.id = dlb.branch_id
WHERE dlb.discount_list_id = $1
ORDER BY b.name;

-- name: UpsertDiscountListBranch :exec
INSERT INTO discount_list_branches (branch_id, discount_list_id)
VALUES ($1, $2)
ON CONFLICT (branch_id)
DO UPDATE SET discount_list_id = EXCLUDED.discount_list_id;

-- name: DeleteDiscountListBranchesForList :exec
DELETE FROM discount_list_branches
WHERE discount_list_id = $1;

-- name: DeleteDiscountListBranch :exec
DELETE FROM discount_list_branches
WHERE branch_id = $1;
