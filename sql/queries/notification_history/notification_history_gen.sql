-- name: Create :one
INSERT INTO notification_history (id, provider_id, incident_id, message, status, error_message)
	VALUES (?, ?, ?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM notification_history WHERE id=?;

-- name: GetByID :one
SELECT * FROM notification_history WHERE id=? LIMIT 1;

