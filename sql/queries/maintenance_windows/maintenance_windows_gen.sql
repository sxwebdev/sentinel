-- name: Create :one
INSERT INTO maintenance_windows (id, name, start_at, end_at, reason)
	VALUES (?, ?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM maintenance_windows WHERE id=?;

-- name: GetAll :many
SELECT * FROM maintenance_windows;

-- name: GetByID :one
SELECT * FROM maintenance_windows WHERE id=? LIMIT 1;

