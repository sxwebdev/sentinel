-- name: GetAllByResourceID :many
SELECT * FROM resource_agents WHERE resource_id=?;

-- name: DeleteByResourceID :exec
DELETE FROM resource_agents WHERE resource_id=?;

-- name: DeleteByResourceIDAndAgentID :exec
DELETE FROM resource_agents WHERE resource_id=? AND agent_id=?;

-- name: ExistsByResourceIDAndAgentID :one
SELECT EXISTS (SELECT 1 FROM resource_agents WHERE resource_id=? AND agent_id=? LIMIT 1);

-- name: CountByResourceID :one
SELECT COUNT(*) FROM resource_agents WHERE resource_id=?;
