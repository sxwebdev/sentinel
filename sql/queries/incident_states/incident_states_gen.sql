-- name: Create :one
INSERT INTO incident_states (id, incident_id, status, level)
	VALUES (?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM incident_states WHERE id=?;

-- name: GetAll :many
SELECT * FROM incident_states;

-- name: GetByID :one
SELECT * FROM incident_states WHERE id=? LIMIT 1;

