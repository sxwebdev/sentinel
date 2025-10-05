-- name: Create :one
INSERT INTO monitor_revisions (id, monitor_id, config, content_hash_uint64, is_active)
	VALUES (?, ?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM monitor_revisions WHERE id=?;

-- name: GetAll :many
SELECT * FROM monitor_revisions;

-- name: GetByID :one
SELECT * FROM monitor_revisions WHERE id=? LIMIT 1;

