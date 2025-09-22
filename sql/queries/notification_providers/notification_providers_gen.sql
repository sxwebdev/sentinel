-- name: Create :one
INSERT INTO notification_providers (id, provider_type, config, is_enabled)
	VALUES (?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM notification_providers WHERE id=?;

-- name: GetByID :one
SELECT * FROM notification_providers WHERE id=? LIMIT 1;

