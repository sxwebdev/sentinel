-- name: Create :one
INSERT INTO agents (id, name, description, token_ct, token_nonce, token_hint, last_assignment_rev, tags, config)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM agents WHERE id=?;

-- name: GetByID :one
SELECT * FROM agents WHERE id=? LIMIT 1;

