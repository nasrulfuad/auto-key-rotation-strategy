# JWKS key rotation simulation

Two Go services that reproduce a streaming provider partner integration built on JWT and JWKS, scaled down from days to minutes so a full key rotation can be watched in one sitting.

| App | Plays | Module | Port |
|---|---|---|---|
| `my_app` | the partner (us) | `github.com/nasrulfuad/jwks_key_rotation/my_app` | 8081 |
| `partner_app` | Provider (Entitlement API owner) | `github.com/nasrulfuad/jwks_key_rotation/partner_app` | 8082 |

Trust goes both ways. Each app owns an RSA signing key set, rotates it on its own schedule, publishes the public half at `/.well-known/jwks.json`, and verifies the other app's tokens through a cached copy of the other app's JWKS.

- `my_app` signs the bearer token of every activation and termination it sends to `partner_app` (`aud` = `set.entitlement...`).
- `partner_app` signs every activation notification webhook it sends back to `my_app` (`aud` = `activation-notification...`).

## Integration flows

### my_app → partner_app: activation and termination

`my_app` signs a short-lived JWT with its active key and calls the Entitlement API. `partner_app` verifies the token against its cached copy of the `my_app` JWKS, which the `refresh_peer_signing_keys` job refreshes every `PEER_JWKS_POLL_INTERVAL`.

```mermaid
sequenceDiagram
    autonumber
    actor Client
    participant MA as my_app
    participant MADB as my_app DB
    participant PA as partner_app
    participant Cache as partner_app JWKS cache<br/>(copy of my_app JWKS)
    participant PADB as partner_app DB

    loop every PEER_JWKS_POLL_INTERVAL
        PA->>MA: GET /.well-known/jwks.json
        MA-->>PA: pending + active + inactive public keys
        PA->>Cache: replace cached keys
    end

    Client->>MA: POST /activations (or /terminations)
    MA->>MADB: find active signing key
    MA->>MA: sign JWT (kid = active key, aud = set.entitlement..., iss, provider, jti, exp = TOKEN_TTL)
    MA->>PA: PUT /subjects/{subject_id}/entitlements/{entitlement_id}<br/>(or POST .../terminate), Authorization: Bearer JWT
    PA->>Cache: find public key by kid
    alt kid not cached and PEER_JWKS_REFRESH_ON_UNKNOWN_KEY_ID=true
        PA->>MA: GET /.well-known/jwks.json (at most once per PEER_JWKS_REFRESH_MIN_INTERVAL)
        MA-->>PA: public keys
        PA->>Cache: replace cached keys
    end
    PA->>PA: verify RS256 signature, iss, aud, exp, provider
    alt token valid
        PA->>PADB: upsert entitlement + queue activation notification (one transaction)
        PA-->>MA: 200 entitlement
        MA->>MADB: upsert subscription with signed_with_key_id = kid
        MA-->>Client: 200 subscription
    else token invalid or kid unknown
        PA-->>MA: 401 TOKEN_*
        MA-->>Client: 502 PARTNER_REQUEST_REJECTED
    end
```

### partner_app → my_app: activation notification webhook

Every entitlement change queues a notification in the same transaction. The `deliver_activation_notifications` job sends due notifications every `ACTIVATION_NOTIFICATION_INTERVAL`, signed with the active `partner_app` key, and retries with backoff until the maximum attempts are reached.

```mermaid
sequenceDiagram
    autonumber
    participant PA as partner_app
    participant PADB as partner_app DB
    participant MA as my_app
    participant Cache as my_app JWKS cache<br/>(copy of partner_app JWKS)
    participant MADB as my_app DB

    loop every PEER_JWKS_POLL_INTERVAL
        MA->>PA: GET /.well-known/jwks.json
        PA-->>MA: pending + active + inactive public keys
        MA->>Cache: replace cached keys
    end

    loop every ACTIVATION_NOTIFICATION_INTERVAL
        PA->>PADB: list due pending notifications
        loop each notification
            PA->>PA: sign JWT (kid = active key, aud = activation-notification..., sub = subject_id)
            PA->>MA: POST /activation-notifications, Authorization: Bearer JWT
            MA->>Cache: find public key by kid (refresh on unknown kid)
            MA->>MA: verify signature, iss, aud, exp, provider and payload
            alt token valid
                MA->>MADB: store notification (duplicates by notification_id ignored)
                MA-->>PA: 2xx
                PA->>PADB: mark sent
            else token invalid or my_app unreachable
                MA-->>PA: 401 TOKEN_* / network error
                PA->>PADB: record failed attempt, next_attempt_at = now + backoff<br/>(status failed after the maximum attempts)
            end
        end
    end
```

## Key lifecycle

| Phase | Default | In JWKS | Signs | Its tokens |
|---|---|---|---|---|
| pending (newly introduced) | `KEY_PENDING_DURATION=5m` | yes | no | — |
| active (primary signing key) | `KEY_ACTIVE_DURATION=20m` | yes | **yes** | valid |
| inactive | `KEY_INACTIVE_DURATION=10m` | yes | no | **still valid** |
| deleted | — | no | no | **rejected** |

The next key is created `KEY_PENDING_DURATION` before the current key stops signing and becomes active exactly when the current key ends, so there is never a gap and never two signing keys. The status is derived from four timestamps (`introduced_at`, `active_from`, `active_until`, `delete_after`); there is no status column. `TOKEN_TTL` must not exceed `KEY_INACTIVE_DURATION`, which guarantees a token signed at the last second of the active phase verifies until it expires.

```mermaid
gantt
    dateFormat HH:mm
    axisFormat %H:%M
    section Key 1
    Active    :crit, k1a, 00:00, 20m
    Inactive  :k1i, after k1a, 10m
    section Key 2
    Pending   :k2p, 00:15, 5m
    Active    :crit, k2a, after k2p, 20m
    Inactive  :k2i, after k2a, 10m
    section Key 3
    Pending   :k3p, 00:35, 5m
    Active    :crit, k3a, after k3p, 20m
```

## Quick start

```
make up        # postgres + both apps, fast profile from .env.example (30s / 2m / 1m)
make dev       # same, but both apps run under air: saving a .go/.sql/.json file rebuilds and restarts that app
make demo      # activate / terminate every 3s and print which kid signed each side
make logs      # watch "signing key created", "signing key activated", "signing key retired", "peer signing key discovered", "signing key deleted"
make down-v    # remove everything including the database volume
```

Every log line carries `source`: `api` lines belong to one HTTP request and share its `log_id` (also returned as the `X-Log-ID` response header); `tick` lines belong to one scheduler run and carry `job` plus that run's `log_id`.

Defaults without `.env` are the 5m / 20m / 10m production-like schedule. `PARTNER_APP_KEY_PENDING_DURATION`, `PARTNER_APP_KEY_ACTIVE_DURATION` and `PARTNER_APP_KEY_INACTIVE_DURATION` give `partner_app` its own schedule; left empty, it uses the shared `KEY_*_DURATION` values. `podman compose` is used when podman is installed, `docker compose` otherwise.

API docs (Swagger UI over the generated OpenAPI 3.1 spec, served only when `ENABLE_API_DOCS=true`, which compose sets by default):

- my_app: http://localhost:8081/docs (spec at `/openapi.json`)
- partner_app: http://localhost:8082/docs

The spec is generated from swag annotations on the handlers (`make docs`, `make -C <app> docs-check`). "Try it" works directly for `my_app` `/activations`, `/terminations` and `/subscriptions`; the `partner_app` entitlement endpoints need a bearer token signed by `my_app`, which is why every call to the partner goes through an activation or termination in `my_app`.

Manual calls:

```
curl -s -X POST localhost:8081/activations -H 'Content-Type: application/json' \
  -d '{"subject_id":"SUB001","entitlement_id":"ENT001","products":"com.partner.product1","offer":{"free_trial":"NONE"}}'
curl -s -X POST localhost:8081/terminations -H 'Content-Type: application/json' \
  -d '{"subject_id":"SUB001","entitlement_id":"ENT001"}'
curl -s localhost:8081/subscriptions
curl -s localhost:8081/activation-notifications
curl -s localhost:8082/.well-known/jwks.json
curl -s localhost:8082/signing-keys
curl -s localhost:8082/peer-signing-keys
```

To see why the pending phase and polling matter, restart with `PEER_JWKS_POLL_INTERVAL=1h` and `PEER_JWKS_REFRESH_ON_UNKNOWN_KEY_ID=false`: every call fails with `401 TOKEN_SIGNING_KEY_NOT_FOUND` as soon as the peer starts signing with a key this side never fetched.

## Endpoints

`partner_app`

| Method | Path | Auth |
|---|---|---|
| GET | `/health`, `/health/db` | public |
| GET | `/docs`, `/openapi.json` | public, only with `ENABLE_API_DOCS=true` |
| GET | `/.well-known/jwks.json` | public |
| GET | `/signing-keys` | public, debug view of every key and its status |
| GET | `/peer-signing-keys` | public, debug view of the cached `my_app` JWKS |
| PUT | `/subjects/:subject_id/entitlements/:entitlement_id` | bearer JWT from `my_app`, `aud` = `ENTITLEMENT_SET_AUDIENCE`; status `ACTIVE` (also reactivates) |
| POST | `/subjects/:subject_id/entitlements/:entitlement_id/terminate` | bearer JWT from `my_app`, `aud` = `ENTITLEMENT_SET_AUDIENCE`; status `TERMINATED` |
| GET | `/subjects/:subject_id/entitlements/:entitlement_id` | bearer JWT from `my_app`, `aud` = `ENTITLEMENT_GET_AUDIENCE` |
| GET | `/subjects/:subject_id/entitlements` | bearer JWT from `my_app`, `aud` = `ENTITLEMENT_GET_AUDIENCE` |

`my_app`

| Method | Path | Auth |
|---|---|---|
| GET | `/health`, `/health/db`, `/.well-known/jwks.json`, `/signing-keys`, `/peer-signing-keys` | public |
| GET | `/docs`, `/openapi.json` | public, only with `ENABLE_API_DOCS=true` |
| POST | `/activations` | public; signs with the active key, calls `partner_app` PUT entitlement, stores the subscription locally with `signed_with_key_id` |
| POST | `/terminations` | public; signs, calls `partner_app` terminate, stores the subscription locally |
| GET | `/subscriptions` | public, local subscriptions with the `kid` of the last call |
| POST | `/activation-notifications` | bearer JWT from `partner_app`, `aud` = `ACTIVATION_NOTIFICATION_AUDIENCE` |
| GET | `/activation-notifications` | public, shows the `key_id` that signed each received webhook |

Errors use one envelope: `{"error":{"code":"...","message":"..."}}`. Token failures are `401` with `TOKEN_INVALID`, `TOKEN_ISSUER_UNKNOWN`, `TOKEN_SIGNING_KEY_NOT_FOUND` or `TOKEN_CLAIM_INVALID`; the precise reason is logged, never returned.

## Tests

```
make test      # unit tests of both apps
make db        # postgres only
make test-db   # integration tests: full rotation with a fake clock, polling failure, entitlement + webhook flows
make lint      # gofmt, go vet, golangci-lint with depguard ring rules
make docs      # regenerate api/openapi.json of both apps from the swag annotations
```

## Not simulated

HTTPS/TLS and the SSL Labs "A" requirement, IP allowlists, the activation token flow for the provider web app, data reconciliation files, and the "do not sign with a new key until the peer has pulled it" probe (the pending phase plus polling covers it here).
