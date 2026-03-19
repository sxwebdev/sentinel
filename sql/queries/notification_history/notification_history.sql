-- name: GetAllUnsent :many
SELECT
    h.*,
    p.provider_type,
    p.config,
    p.project_id
  FROM notification_history h
  LEFT JOIN notification_providers p ON p.id = h.provider_id
  WHERE
    h.status != 'sent' AND
    p.is_enabled = true
  ORDER BY h.created_at ASC
  LIMIT 100;

-- name: IncrementAttempt :exec
UPDATE notification_history
  SET
    response = ?,
    attempts = attempts + 1,
    error_message = ?,
    last_attempt_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
  WHERE id = ?;

-- name: MarkAsSent :exec
UPDATE notification_history
  SET
    status = 'sent',
    response = ?,
    attempts = attempts + 1,
    error_message = NULL,
    last_attempt_at = CURRENT_TIMESTAMP,
    sent_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
  WHERE id = ?;

-- name: DeleteAllByProjectID :exec
DELETE FROM notification_history
  WHERE provider_id in (
    SELECT id FROM notification_providers WHERE project_id = ?
  );
