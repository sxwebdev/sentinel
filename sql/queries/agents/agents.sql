-- name: Exist :one
SELECT EXISTS (SELECT 1 FROM agents WHERE id=? LIMIT 1);
