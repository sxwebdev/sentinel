-- name: DeleteByServiceID :exec
DELETE FROM service_states WHERE service_id=?;
