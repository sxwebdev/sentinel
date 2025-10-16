-- name: Exist :one
SELECT EXISTS (SELECT 1 FROM agents WHERE id=? LIMIT 1);

-- name: GetByIDAndProjectID :one
SELECT * FROM agents WHERE id=? AND project_id=? LIMIT 1;
