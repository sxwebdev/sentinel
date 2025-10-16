-- name: Create :one
INSERT INTO agents (id, name, description, secret_hash, token_hint, kind, location, tags, config, project_id)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM agents WHERE id=? AND project_id=?;

-- name: GetByID :one
SELECT * FROM agents WHERE id=? LIMIT 1;

