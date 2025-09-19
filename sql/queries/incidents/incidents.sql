-- name: DeleteByServiceID :exec
DELETE FROM incidents WHERE service_id=?;

-- name: StatsByServiceID :one
SELECT
 	COUNT(*) AS total_incidents,
 	SUM(duration) AS total_downtime,
  AVG(duration) AS avg_downtime,
  SUM(CASE WHEN resolved THEN 1 ELSE 0 END) AS resolved_incidents,
  SUM(CASE WHEN NOT resolved THEN 1 ELSE 0 END) AS unresolved_incidents
FROM incidents
WHERE service_id=? AND start_time >= ?;

-- name: GetAllUnresolvedByServiceID :many
SELECT * FROM incidents
WHERE service_id=? AND NOT resolved
ORDER BY start_time DESC;

-- name: ResolveByID :exec
UPDATE incidents
SET
  resolved = TRUE,
  end_time = CURRENT_TIMESTAMP,
  duration = CAST((julianday('now') - julianday(start_time)) * 86400000 AS INTEGER),
  updated_at = CURRENT_TIMESTAMP
WHERE id=? AND NOT resolved;
