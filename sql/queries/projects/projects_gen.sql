-- name: Create :one
INSERT INTO projects (id, name, description, settings)
	VALUES (?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM projects WHERE id=?;

-- name: GetAll :many
SELECT * FROM projects;

-- name: GetByID :one
SELECT * FROM projects WHERE id=? LIMIT 1;

-- name: Update :one
UPDATE projects
	SET name=?, description=?, settings=?, updated_at=CURRENT_TIMESTAMP
	WHERE id=?
	RETURNING *;

