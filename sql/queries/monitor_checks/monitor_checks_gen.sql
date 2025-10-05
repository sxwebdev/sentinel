-- name: Create :one
INSERT INTO monitor_checks (id, project_id, monitor_id, revision_id, resource_id, agent_id, started_at, completed_at, severity, response_time, evaluation, meta)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM monitor_checks WHERE id=?;

-- name: GetAll :many
SELECT * FROM monitor_checks;

-- name: GetByID :one
SELECT * FROM monitor_checks WHERE id=? LIMIT 1;

