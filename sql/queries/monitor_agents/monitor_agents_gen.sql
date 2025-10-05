-- name: Create :one
INSERT INTO monitor_agents (id, project_id, monitor_id, agent_id, created_at)
	VALUES (?, ?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM monitor_agents WHERE id=?;

-- name: GetAll :many
SELECT * FROM monitor_agents;

-- name: GetByID :one
SELECT * FROM monitor_agents WHERE id=? LIMIT 1;

