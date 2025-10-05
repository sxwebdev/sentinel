-- name: Create :one
INSERT INTO notification_providers (id, provider_type, config, is_enabled, project_id)
	VALUES (?, ?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM notification_providers WHERE id=?;

-- name: GetAll :many
SELECT * FROM notification_providers;

-- name: GetByID :one
SELECT * FROM notification_providers WHERE id=? LIMIT 1;

