-- name: Create :one
INSERT INTO incidents (id, service_id, start_time, end_time, error, duration_ns, resolved)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM incidents WHERE id=?;

-- name: GetAll :many
SELECT * FROM incidents;

-- name: GetByID :one
SELECT * FROM incidents WHERE id=? LIMIT 1;

