-- name: Create :one
INSERT INTO agents (id, name, description, secret_hash, token_hint, last_assignment_rev, tags, config, last_connected_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM agents WHERE id=?;

-- name: GetByID :one
SELECT * FROM agents WHERE id=? LIMIT 1;

