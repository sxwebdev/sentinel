-- name: GetAllEnabled :many
SELECT * FROM notification_providers WHERE is_enabled=true;

-- name: Update :exec
UPDATE notification_providers
  SET
    provider_type = ?,
    config = ?,
    is_enabled = ?,
    updated_at = CURRENT_TIMESTAMP
  WHERE id = ?;
