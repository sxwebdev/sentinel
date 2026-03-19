-- name: Create :one
INSERT INTO alerts (id, incident_id, policy_id, status)
	VALUES (?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM alerts WHERE id=?;

-- name: GetAll :many
SELECT * FROM alerts;

-- name: GetByID :one
SELECT * FROM alerts WHERE id=? LIMIT 1;

