-- name: Create :one
INSERT INTO monitor_states (id, project_id, monitor_id, resource_id, agent_id, severity, status, last_check, last_error, consecutive_fails, consecutive_success, total_checks, avg_response_time)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM monitor_states WHERE id=?;

-- name: GetAll :many
SELECT * FROM monitor_states;

-- name: GetByID :one
SELECT * FROM monitor_states WHERE id=? LIMIT 1;

