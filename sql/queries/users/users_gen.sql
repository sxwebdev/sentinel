-- name: Create :one
INSERT INTO users (id, email, password, full_name, role, avatar)
	VALUES (?, ?, ?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM users WHERE id=?;

-- name: GetByID :one
SELECT * FROM users WHERE id=? LIMIT 1;

-- name: Update :one
UPDATE users
	SET email=?, full_name=?, avatar=?, updated_at=CURRENT_TIMESTAMP
	WHERE id=?
	RETURNING *;

