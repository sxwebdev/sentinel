-- name: DeleteByServiceID :exec
DELETE FROM incidents WHERE service_id=?;

-- name: Stats :one
SELECT
 	COUNT(*) AS total_incidents,
 	SUM(duration) AS total_downtime,
  AVG(duration) AS avg_downtime,
  SUM(CASE WHEN resolved_at IS NOT NULL THEN 1 ELSE 0 END) AS resolved_incidents,
  SUM(CASE WHEN resolved_at IS NULL THEN 1 ELSE 0 END) AS unresolved_incidents
FROM incidents;

-- name: StatsByServiceID :one
SELECT
 	COUNT(*) AS total_incidents,
 	SUM(duration) AS total_downtime,
  AVG(duration) AS avg_downtime,
  SUM(CASE WHEN resolved_at IS NOT NULL THEN 1 ELSE 0 END) AS resolved_incidents,
  SUM(CASE WHEN resolved_at IS NULL THEN 1 ELSE 0 END) AS unresolved_incidents,
  ROUND(100.0 - (COALESCE(SUM(duration), 0) * 100.0 / (30 * 24 * 60 * 60 * 1000)), 3) AS uptime_percentage_30d
FROM incidents
WHERE service_id=? AND created_at >= ?;

-- name: GetAllUnresolvedByServiceID :many
SELECT * FROM incidents
WHERE service_id=? AND resolved_at IS NULL
ORDER BY created_at DESC;

-- name: ResolveByID :exec
UPDATE incidents
SET
  resolved_at = CURRENT_TIMESTAMP,
  duration = CAST((julianday('now') - julianday(started_at)) * 86400000 AS INTEGER),
  updated_at = CURRENT_TIMESTAMP
WHERE id=? AND resolved_at IS NULL;
