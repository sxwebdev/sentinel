-- name: Exist :one
SELECT EXISTS (SELECT 1 FROM monitors WHERE id = ? LIMIT 1);

-- name: GetAllMonitorsByAgentID :many
WITH
  target_agent AS (
    SELECT a.*
    FROM agents a
    WHERE a.id = ?
      AND a.is_enabled = 1
  ),

  override_monitors AS (
    SELECT ma.monitor_id
    FROM monitor_agents ma
    JOIN target_agent ta ON ta.id = ma.agent_id
  ),

  monitors_with_override AS (
    SELECT DISTINCT ma.monitor_id
    FROM monitor_agents ma
  ),

  inherited_monitors AS (
    SELECT m.id AS monitor_id
    FROM monitors m
    JOIN resources r        ON r.id = m.resource_id
    JOIN target_agent ta    ON 1=1
    JOIN resource_agents ra ON ra.resource_id = r.id
                           AND ra.agent_id    = ta.id
    WHERE m.is_enabled = 1
      AND ta.is_enabled = 1
      AND r.project_id  = ta.project_id
      AND NOT EXISTS (SELECT 1
                      FROM monitors_with_override o
                      WHERE o.monitor_id = m.id)
  ),

  effective_monitors AS (
    SELECT monitor_id FROM override_monitors
    UNION
    SELECT monitor_id FROM inherited_monitors
  )

SELECT
  m.*,
  mr.id      AS active_revision_id,
  mr.config  AS active_revision_config
FROM effective_monitors em
JOIN monitors m             ON m.id = em.monitor_id
JOIN resources r            ON r.id = m.resource_id
JOIN target_agent ta        ON r.project_id = ta.project_id
LEFT JOIN monitor_revisions mr
       ON mr.monitor_id = m.id AND mr.is_active = 1
ORDER BY m.resource_id, m.name;

-- namg: GetMonitorsForAgents :many
WITH
  enabled_agents AS (
    SELECT a.*
    FROM agents a
    WHERE a.is_enabled = 1
  ),

  override_pairs AS (
    SELECT
      ma.agent_id,
      ma.monitor_id
    FROM monitor_agents ma
    JOIN enabled_agents a ON a.id = ma.agent_id
  ),

  monitors_with_override AS (
    SELECT DISTINCT monitor_id
    FROM monitor_agents
  ),

  inherited_pairs AS (
    SELECT
      ra.agent_id,
      m.id AS monitor_id
    FROM resource_agents ra
    JOIN enabled_agents a ON a.id = ra.agent_id
    JOIN resources r      ON r.id = ra.resource_id
    JOIN monitors  m      ON m.resource_id = r.id
    WHERE m.is_enabled = 1
      AND r.project_id = a.project_id
      AND NOT EXISTS (
            SELECT 1 FROM monitors_with_override o
            WHERE o.monitor_id = m.id
          )
  )

SELECT DISTINCT
  p.agent_id,
  p.monitor_id
FROM (
  SELECT * FROM override_pairs
  UNION ALL
  SELECT * FROM inherited_pairs
) AS p
ORDER BY p.agent_id, p.monitor_id;
