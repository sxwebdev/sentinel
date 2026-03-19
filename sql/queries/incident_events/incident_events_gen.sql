-- name: Create :one
INSERT INTO incident_events (id, incident_id, type, payload)
	VALUES (?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM incident_events WHERE id=?;

-- name: GetAll :many
SELECT * FROM incident_events;

-- name: GetByID :one
SELECT * FROM incident_events WHERE id=? LIMIT 1;

