-- name: ListFacilities :many
SELECT id, name, region, town, type, beds, lifecycle, health,
       manager_name, payer_nhis, payer_cash_momo, payer_private,
       latitude, longitude
FROM facilities
ORDER BY name;

-- name: CreateFacility :exec
INSERT INTO facilities (
    id, name, region, town, type, beds, lifecycle, health,
    manager_name, payer_nhis, payer_cash_momo, payer_private, latitude, longitude
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
);
