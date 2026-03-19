-- name: Create :one
INSERT INTO alert_policies (id, name, rules, is_enabled, project_id)
	VALUES (?, ?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM alert_policies WHERE id=?;

-- name: GetAll :many
SELECT * FROM alert_policies;

-- name: GetByID :one
SELECT * FROM alert_policies WHERE id=? LIMIT 1;

