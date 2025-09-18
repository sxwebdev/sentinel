-- name: Exist :one
SELECT EXISTS (SELECT 1 FROM services WHERE id = ? LIMIT 1);
