-- name: GetPriceListBranch :one
SELECT * FROM price_list_branches
WHERE branch_id = $1 LIMIT 1;

-- name: ListPriceListBranchIDs :many
SELECT branch_id FROM price_list_branches
WHERE price_list_id = $1
ORDER BY branch_id;

-- name: ListBranchesForPriceList :many
SELECT b.id, b.name, b.code
FROM price_list_branches plb
JOIN branches b ON b.id = plb.branch_id
WHERE plb.price_list_id = $1
ORDER BY b.name;

-- name: UpsertPriceListBranch :exec
INSERT INTO price_list_branches (branch_id, price_list_id)
VALUES ($1, $2)
ON CONFLICT (branch_id)
DO UPDATE SET price_list_id = EXCLUDED.price_list_id;

-- name: DeletePriceListBranchesForList :exec
DELETE FROM price_list_branches
WHERE price_list_id = $1;

-- name: DeletePriceListBranch :exec
DELETE FROM price_list_branches
WHERE branch_id = $1;
