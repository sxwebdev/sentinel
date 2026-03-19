-- name: Create :one
INSERT INTO notification_history (id, alert_id, provider_id, message)
	VALUES (?, ?, ?, ?)
	RETURNING *;

