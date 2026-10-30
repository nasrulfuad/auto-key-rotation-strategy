-- name: UpsertSubscription :one
INSERT INTO subscription (
    id, subject_id, entitlement_id, products, free_trial, status, signed_with_key_id, activated_at, terminated_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, sqlc.arg(now)::TIMESTAMPTZ, sqlc.arg(now)::TIMESTAMPTZ
)
ON CONFLICT (subject_id, entitlement_id) DO UPDATE
SET products = EXCLUDED.products,
    free_trial = EXCLUDED.free_trial,
    status = EXCLUDED.status,
    signed_with_key_id = EXCLUDED.signed_with_key_id,
    activated_at = COALESCE(EXCLUDED.activated_at, subscription.activated_at),
    terminated_at = EXCLUDED.terminated_at,
    updated_at = EXCLUDED.updated_at
RETURNING id, subject_id, entitlement_id, products, free_trial, status, signed_with_key_id, activated_at, terminated_at, created_at, updated_at;

-- name: ListAllSubscriptions :many
SELECT id, subject_id, entitlement_id, products, free_trial, status, signed_with_key_id, activated_at, terminated_at, created_at, updated_at
FROM subscription
ORDER BY updated_at DESC, id
LIMIT sqlc.arg(row_limit);
