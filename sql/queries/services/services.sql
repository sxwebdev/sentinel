-- name: Exist :one
SELECT EXISTS (SELECT 1 FROM services WHERE id = ? LIMIT 1);

-- name: GetAllEnabled :many
SELECT * FROM services WHERE is_enabled=TRUE ORDER BY name;

-- name: GetAllEnabledByAgentID :many
SELECT s.* FROM services s
  JOIN services_agents ags ON s.id = ags.service_id
  WHERE ags.agent_id = ? AND s.is_enabled=TRUE
  ORDER BY s.created_at DESC;

-- name: GetAllEnabledWithoutAgents :many
SELECT * FROM services s
  WHERE s.is_enabled=TRUE AND s.id NOT IN (
    SELECT service_id FROM services_agents sa WHERE sa.service_id = s.id
  )
  ORDER BY s.created_at DESC;
