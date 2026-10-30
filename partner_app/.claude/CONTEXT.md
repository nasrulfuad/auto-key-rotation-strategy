# Project context — partner_app

partner_app is one half of the JWKS key rotation simulation at the repository root. It plays the provider. It owns the Entitlement API, verifies my_app's bearer tokens through my_app's JWKS, and signs the activation notification webhooks it delivers to my_app.

Module path: `github.com/nasrulfuad/jwks_key_rotation/partner_app`. Working rules (English only, no comments, clean architecture, naming standard) live in the root `CLAUDE.md`; the ring rules are in `rules/rules.md` and enforced by depguard in `.golangci.yml`.

## Where the code for X lives

| Layer | Path | Rule |
|---|---|---|
| Entities, schedule, token rules | `internal/domain/` | stdlib + `google/uuid` only; no `net/http`, `database/sql`, `crypto/rsa` |
| Interactors + driven ports | `internal/usecase/<feature>/`, `internal/usecase/port/` | sees `domain` only |
| Route table, controllers, peer auth middleware | `internal/adapter/controller/http/` | `router.go` is the whole HTTP surface |
| JSON bodies and the error envelope | `internal/adapter/presenter/` | the only place that knows HTTP shapes |
| Doc-only mirrors of the JSON bodies | `internal/adapter/apidoc/` | kept honest by `apidoc_contract_test.go` in `presenter` and `controller/http` |
| Outbound HTTP to the peer | `internal/adapter/gateway/http/` | |
| SQL repositories | `internal/adapter/repository/postgres/` | the only importer of `internal/store` (plus `di`) |
| SQL / generated code | `query/*.sql` → `internal/store/` | sqlc; `make sqlc` |
| Schema | `migrations/*.sql` | Up only; sqlc reads it as the schema |
| RSA, AES-GCM, JWT, JWKS cache, scheduler | `internal/infra/{rsakey,sealer,jwt,jwks,scheduler}` | |
| Wiring | `internal/di/container.go` | `NewWithDB` is reused by the integration tests |
| Integration tests | `internal/integration/` | build tag `integration`, needs `TEST_DATABASE_ADMIN_URL` |

## Signing key lifecycle

- `domain.SigningKeySchedule.PlanNext(latest, now)` is the whole rotation rule: no key → bootstrap (active now); latest key already ended → recovery (active now, logged as WARN because peers had no pending phase to fetch it); `now >= latest.active_until - pending` → rotation (`active_from = latest.active_until`). Windows are therefore contiguous and exactly one key can sign at any instant.
- Status is derived (`SigningKey.StatusAt(now)`): pending → active → inactive → expired. There is no status column.
- Jobs: `ensure_signing_keys` and `delete_expired_signing_keys` run every `KEY_ROTATION_CHECK_INTERVAL` under a PostgreSQL advisory lock (`JobLockRepository`); `signing_key_active_from_key` UNIQUE is the second safety net. `main` runs `ensure_signing_keys` once before serving.
- `kid` is the RFC 7638 SHA-256 thumbprint (base64url, no padding), so a kid can never be reused.
- Private keys are PKCS#8, sealed with AES-256-GCM (`KEY_ENCRYPTION_KEY`), returned only through `SigningKeyReader.FindPrivateKey` as `domain.Secret`, which refuses to print or marshal.
- `/.well-known/jwks.json` lists every key with `introduced_at <= now < delete_after`; deleted keys disappear and their tokens fail with `TOKEN_SIGNING_KEY_NOT_FOUND`.

## Verifying peer tokens

`infra/jwt.Verifier` + `infra/jwks.RemoteKeySet`:

1. `jwt.WithValidMethods` pins `PEER_KEY_ALGORITHM` (blocks `alg: none` and HMAC confusion).
2. `iss` must equal `PEER_ISSUER` before any key lookup.
3. Key by `kid` from the cache; an unknown kid triggers one refresh (singleflight, at most once per `PEER_JWKS_REFRESH_MIN_INTERVAL`, or after 1s when the last fetch failed) when `PEER_JWKS_REFRESH_ON_UNKNOWN_KEY_ID` is true; otherwise every cached key is tried.
4. `domain.TokenExpectation.Check`: `aud` for the route, `exp` and `iat` with `TOKEN_LEEWAY`, lifetime at most 24h.
5. A failed refresh keeps the cached keys. `refresh_peer_signing_keys` polls every `PEER_JWKS_POLL_INTERVAL`.

## Features

- `entitlement`: `UpsertEntitlement` (put_entitlement; status `ACTIVE`, clears `terminated_at`, so it also reactivates), `TerminateEntitlement` (`POST .../terminate`, set audience; status `TERMINATED`, `terminated_at` kept on repeats, 404 when it never existed), `GetEntitlement` (get_entitlement), `ListEntitlements`. The token's `provider` must equal `PEER_PROVIDER` and scopes every lookup. The token's `iat` is stored as `last_updated_at`; an upsert whose `iat` is older than the stored value is `ENTITLEMENT_STALE_UPDATE` (enforced in SQL by `ON CONFLICT ... WHERE last_updated_at <= EXCLUDED.last_updated_at`).
- `activationnotification`: every upsert and termination inserts an `activation_notification` outbox row (`entitlement_upserted` / `entitlement_terminated`) in the same transaction; `DeliverActivationNotifications` (scheduler job) signs a token per row and POSTs it to `ACTIVATION_NOTIFICATION_URL`, with exponential backoff and `failed` after `ACTIVATION_NOTIFICATION_MAX_ATTEMPTS`.
- Tables: `signing_key`, `entitlement`, `activation_notification` (FK `entitlement_row_id`, named so it is not confused with the business `entitlement_id`).

## Configuration

See `internal/infra/config/config.go` for every variable and default, and the root `.env.example` for the fast demo profile. Rejected at boot: `TOKEN_TTL` above `KEY_INACTIVE_DURATION` or 24h, non-positive durations, `KEY_ROTATION_CHECK_INTERVAL >= KEY_PENDING_DURATION`, RSA below 2048 bits, an algorithm outside RS256/RS384/RS512, and a `KEY_ENCRYPTION_KEY` that is not 32 base64 bytes. `PEER_JWKS_POLL_INTERVAL >= KEY_PENDING_DURATION` is only a warning.

## API contract

OpenAPI 3.1 generated by swag v2 (`go tool swag`, pinned in `go.mod`) from the annotations on every handler plus the global block on `cmd/server/main.go`. `go generate .` (`make docs`) writes `api/openapi.json`, then `cmd/fixopenapi` collapses swag's `oneOf [object, $ref]` request bodies and marks nullable fields. The spec is embedded (`OpenAPIFS` in `embed.go`) and served at `GET /openapi.json` plus `GET /docs` (Swagger UI) only when `ENABLE_API_DOCS=true` (default false). Gates: `apidoc_contract_test.go` (mirror ↔ real DTO shapes), `openapi_contract_test.go` (every route in `Router.Table()` is documented and nothing else), `make docs-check` (committed spec matches the annotations). Swag annotations are the one sanctioned exception to the no-comments rule, as in `pricing/backend`.

## Commands

```
make generate   sqlc + mocks + OpenAPI spec
make docs       OpenAPI spec only
make lint       gofmt, vet, golangci-lint (also the integration package)
make test       unit tests
make test-db    integration tests (root `make db` first)
make verify     lint + test + test-db
```
