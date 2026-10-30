-- name: InsertSigningKey :one
INSERT INTO signing_key (
    id, key_id, algorithm, public_key_modulus, public_key_exponent, private_key_ciphertext, private_key_nonce,
    introduced_at, active_from, active_until, delete_after
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING id, key_id, algorithm, public_key_modulus, public_key_exponent, introduced_at, active_from, active_until, delete_after, created_at;

-- name: FindLatestSigningKey :one
SELECT id, key_id, algorithm, public_key_modulus, public_key_exponent, introduced_at, active_from, active_until, delete_after, created_at
FROM signing_key
ORDER BY active_from DESC
LIMIT 1;

-- name: FindActiveSigningKey :one
SELECT id, key_id, algorithm, public_key_modulus, public_key_exponent, introduced_at, active_from, active_until, delete_after, created_at
FROM signing_key
WHERE active_from <= sqlc.arg(now)::TIMESTAMPTZ AND active_until > sqlc.arg(now)::TIMESTAMPTZ
ORDER BY active_from DESC
LIMIT 1;

-- name: FindSigningKeyPrivateKeyByKeyID :one
SELECT private_key_ciphertext, private_key_nonce
FROM signing_key
WHERE key_id = $1;

-- name: ListAllPublishedSigningKeys :many
SELECT id, key_id, algorithm, public_key_modulus, public_key_exponent, introduced_at, active_from, active_until, delete_after, created_at
FROM signing_key
WHERE introduced_at <= sqlc.arg(now)::TIMESTAMPTZ AND delete_after > sqlc.arg(now)::TIMESTAMPTZ
ORDER BY active_from;

-- name: ListAllSigningKeys :many
SELECT id, key_id, algorithm, public_key_modulus, public_key_exponent, introduced_at, active_from, active_until, delete_after, created_at
FROM signing_key
ORDER BY active_from;

-- name: DeleteExpiredSigningKeys :many
DELETE FROM signing_key
WHERE delete_after <= sqlc.arg(now)::TIMESTAMPTZ
RETURNING key_id;
