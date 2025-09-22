-- name: Create :one
INSERT INTO notification_history (id, provider_id, incident_id, message, error_message)
	VALUES (?, ?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM notification_history WHERE id=?;

