-- name: Create :one
INSERT INTO monitors (id, project_id, resource_id, name, kind, is_enabled, tags)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM monitors WHERE id=?;

-- name: GetAll :many
SELECT * FROM monitors;

-- name: GetByID :one
SELECT * FROM monitors WHERE id=? LIMIT 1;

