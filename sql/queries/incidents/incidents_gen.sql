-- name: Create :one
INSERT INTO incidents (id, project_id, origin, monitor_id, resource_id, agent_id, kind, status, severity, summary, first_seen_at, last_seen_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM incidents WHERE id=?;

-- name: GetByID :one
SELECT * FROM incidents WHERE id=? LIMIT 1;

