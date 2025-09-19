-- name: Exist :one
SELECT EXISTS (SELECT 1 FROM services WHERE id = ? LIMIT 1);

-- name: GetAllEnabled :many
SELECT * FROM services WHERE is_enabled=TRUE ORDER BY name;
