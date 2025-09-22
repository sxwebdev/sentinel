-- name: Create :one
INSERT INTO notification_history (id, provider_id, service_id, incident_id, message)
	VALUES (?, ?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM notification_history WHERE id=?;

