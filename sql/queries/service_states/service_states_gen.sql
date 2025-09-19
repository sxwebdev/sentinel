-- name: Create :one
INSERT INTO service_states (id, service_id, status, last_check, next_check, last_error, consecutive_fails, consecutive_success, total_checks, response_time)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM service_states WHERE id=?;

-- name: GetAll :many
SELECT * FROM service_states;

-- name: GetByID :one
SELECT * FROM service_states WHERE id=? LIMIT 1;

