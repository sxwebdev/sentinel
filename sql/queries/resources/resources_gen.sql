-- name: Create :one
INSERT INTO resources (id, project_id, name, description, kind, tags, payload)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM resources WHERE id=?;

-- name: GetAll :many
SELECT * FROM resources WHERE project_id=?;

-- name: GetByID :one
SELECT * FROM resources WHERE id=? LIMIT 1;

