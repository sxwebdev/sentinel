-- name: DeleteByServiceID :exec
DELETE FROM incidents WHERE service_id=?;

-- name: StatsByServiceID :one
SELECT
 	COUNT(*) AS total_incidents,
 	SUM(duration_ns) AS total_downtime,
  AVG(duration_ns) AS avg_downtime,
  SUM(CASE WHEN resolved THEN 1 ELSE 0 END) AS resolved_incidents,
  SUM(CASE WHEN NOT resolved THEN 1 ELSE 0 END) AS unresolved_incidents
FROM incidents
WHERE service_id=? AND start_time >= ?;
