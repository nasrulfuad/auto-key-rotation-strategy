-- name: LockJob :one
SELECT pg_try_advisory_xact_lock(hashtext(sqlc.arg(job_name)::TEXT))::BOOLEAN AS acquired;

-- name: Ping :one
SELECT 1::INT AS ok;
