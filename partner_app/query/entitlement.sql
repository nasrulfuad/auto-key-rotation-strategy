-- name: UpsertEntitlement :one
INSERT INTO entitlement (
    id, provider, subject_id, entitlement_id, products, free_trial, last_updated_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, sqlc.arg(now)::TIMESTAMPTZ, sqlc.arg(now)::TIMESTAMPTZ
)
ON CONFLICT (provider, subject_id, entitlement_id) DO UPDATE
SET products = EXCLUDED.products,
    free_trial = EXCLUDED.free_trial,
    status = 'ACTIVE',
    terminated_at = NULL,
    last_updated_at = EXCLUDED.last_updated_at,
    updated_at = EXCLUDED.updated_at
WHERE entitlement.last_updated_at <= EXCLUDED.last_updated_at
RETURNING id, provider, subject_id, entitlement_id, products, free_trial, status, terminated_at, last_updated_at, created_at, updated_at;

-- name: FindEntitlementBySubjectAndEntitlement :one
SELECT id, provider, subject_id, entitlement_id, products, free_trial, status, terminated_at, last_updated_at, created_at, updated_at
FROM entitlement
WHERE provider = $1 AND subject_id = $2 AND entitlement_id = $3;

-- name: ListAllEntitlementsBySubject :many
SELECT id, provider, subject_id, entitlement_id, products, free_trial, status, terminated_at, last_updated_at, created_at, updated_at
FROM entitlement
WHERE provider = $1 AND subject_id = $2
ORDER BY entitlement_id;

-- name: UpdateEntitlementTermination :one
UPDATE entitlement
SET status = 'TERMINATED',
    terminated_at = COALESCE(terminated_at, sqlc.arg(now)::TIMESTAMPTZ),
    last_updated_at = sqlc.arg(last_updated_at)::TIMESTAMPTZ,
    updated_at = sqlc.arg(now)::TIMESTAMPTZ
WHERE provider = sqlc.arg(provider) AND subject_id = sqlc.arg(subject_id) AND entitlement_id = sqlc.arg(entitlement_id)
  AND last_updated_at <= sqlc.arg(last_updated_at)::TIMESTAMPTZ
RETURNING id, provider, subject_id, entitlement_id, products, free_trial, status, terminated_at, last_updated_at, created_at, updated_at;

-- name: EntitlementExists :one
SELECT EXISTS (
    SELECT 1 FROM entitlement WHERE provider = $1 AND subject_id = $2 AND entitlement_id = $3
)::BOOLEAN AS found;
