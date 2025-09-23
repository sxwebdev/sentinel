-- name: Create :one
INSERT INTO services (id, name, protocol, interval, timeout, retries, tags, config, is_enabled, is_notifications_enabled)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM services WHERE id=?;

-- name: GetByID :one
SELECT * FROM services WHERE id=? LIMIT 1;

