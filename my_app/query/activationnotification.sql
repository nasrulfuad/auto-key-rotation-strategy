-- name: InsertActivationNotification :execrows
INSERT INTO activation_notification (
    id, notification_id, key_id, issuer, event, provider, subject_id, entitlement_id, products, free_trial, received_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
ON CONFLICT (notification_id) DO NOTHING;

-- name: ListAllActivationNotifications :many
SELECT id, notification_id, key_id, issuer, event, provider, subject_id, entitlement_id, products, free_trial, received_at
FROM activation_notification
ORDER BY received_at DESC, id
LIMIT sqlc.arg(row_limit);
