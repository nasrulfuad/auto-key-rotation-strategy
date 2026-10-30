-- name: InsertActivationNotification :exec
INSERT INTO activation_notification (id, entitlement_row_id, event, status, attempts, next_attempt_at, created_at)
VALUES ($1, $2, $3, 'pending', 0, sqlc.arg(now)::TIMESTAMPTZ, sqlc.arg(now)::TIMESTAMPTZ);

-- name: ListAllDueActivationNotifications :many
SELECT
    activation_notification.id,
    activation_notification.event,
    activation_notification.status,
    activation_notification.attempts,
    activation_notification.next_attempt_at,
    activation_notification.created_at,
    entitlement.provider,
    entitlement.subject_id,
    entitlement.entitlement_id,
    entitlement.products,
    entitlement.free_trial
FROM activation_notification
JOIN entitlement ON entitlement.id = activation_notification.entitlement_row_id
WHERE activation_notification.status = 'pending' AND activation_notification.next_attempt_at <= sqlc.arg(now)::TIMESTAMPTZ
ORDER BY activation_notification.next_attempt_at, activation_notification.id
LIMIT sqlc.arg(row_limit);

-- name: UpdateActivationNotificationSent :exec
UPDATE activation_notification
SET status = 'sent', sent_at = sqlc.arg(sent_at)::TIMESTAMPTZ, last_error = NULL
WHERE id = $1;

-- name: UpdateActivationNotificationAttemptFailed :exec
UPDATE activation_notification
SET status = $2, attempts = $3, next_attempt_at = $4, last_error = $5
WHERE id = $1;
