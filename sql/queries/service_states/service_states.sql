-- name: GetByServiceID :one
SELECT * FROM service_states WHERE service_id=? LIMIT 1;

-- name: DeleteByServiceID :exec
DELETE FROM service_states WHERE service_id=?;

-- name: Stats :one
SELECT
 	COUNT(*) AS total_services,
 	SUM(CASE WHEN status='up' THEN 1 ELSE 0 END) AS services_up,
 	SUM(CASE WHEN status='down' THEN 1 ELSE 0 END) AS services_down,
 	SUM(CASE WHEN status='unknown' THEN 1 ELSE 0 END) AS services_unknown,
 	AVG(response_time) AS avg_response_time,
 	SUM(total_checks) AS total_checks           
FROM service_states;
