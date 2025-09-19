-- name: Create :one
INSERT INTO incidents (id, service_id, start_time, error)
	VALUES (?, ?, CURRENT_TIMESTAMP, sqlc.arg(incident_error))
	RETURNING *;

-- name: Delete :exec
DELETE FROM incidents WHERE id=?;

-- name: GetByID :one
SELECT * FROM incidents WHERE id=? LIMIT 1;

