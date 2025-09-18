-- name: DeleteByServiceID :exec
DELETE FROM incidents WHERE service_id=?;
