-- name: Create :one
INSERT INTO resources (id, project_id, name, description, tags, payload)
	VALUES (?, ?, ?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM resources WHERE id=?;

-- name: GetAll :many
SELECT * FROM resources;

-- name: GetByID :one
SELECT * FROM resources WHERE id=? LIMIT 1;

