-- name: GetByServiceID :one
SELECT * FROM service_states WHERE service_id=? LIMIT 1;

-- name: DeleteByServiceID :exec
DELETE FROM service_states WHERE service_id=?;
