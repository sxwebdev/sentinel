-- name: Create :one
INSERT INTO resource_agents (id, resource_id, agent_id, project_id, created_at)
	VALUES (?, ?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM resource_agents WHERE id=?;

-- name: GetAll :many
SELECT * FROM resource_agents;

-- name: GetByID :one
SELECT * FROM resource_agents WHERE id=? LIMIT 1;

