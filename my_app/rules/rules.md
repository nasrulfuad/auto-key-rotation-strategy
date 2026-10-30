# Clean Architecture Rules

These rules apply to every change in `my_app`. They mirror `pricing/backend/rules/rules.md`, reduced to what this service contains.

## 1. Rings

```
cmd/server          entry point
  internal/di       composition root, the only place that sees the whole graph
    internal/infra      config, echo, postgres, clock, logging, scheduler, rsakey, sealer, jwt, jwks
    internal/adapter    controller/http, presenter, gateway/http, repository/postgres
      internal/store    sqlc-generated code
    internal/usecase    interactors + port interfaces
    internal/domain     entities, value objects, rules
```

Dependencies point inward only: `infra/adapter -> usecase -> domain`. depguard in `.golangci.yml` fails the lint when they do not.

## 2. Domain (`internal/domain`)

- Standard library and `github.com/google/uuid` only. `net/http`, `database/sql` and `crypto/rsa` are denied: key material is stored as bytes (`PublicKey`, `Secret`) and converted by `infra/rsakey`.
- Holds every rule that can be expressed without I/O: the rotation schedule (`SigningKeySchedule.PlanNext`), key status, token expectations, identifier validation, retry backoff.
- Functions that depend on time take `now` as an argument.

## 3. Use cases (`internal/usecase`)

- One interactor per file, `<Verb><Object>Interactor` with a single `Execute(ctx, in)`.
- Dependencies are port interfaces from `internal/usecase/port`, injected through the constructor and named after the interface.
- `Clock.Now()` is called once per `Execute` and passed down.
- May not import Echo, `database/sql`, `golang-jwt`, `internal/store`, `internal/adapter` or `internal/infra`.
- A dependency implemented by another use case (the token `Signer`) is consumed through a small interface declared in the consuming package (`TokenSigner`).

## 4. Ports (`internal/usecase/port`)

- `<Entity>Reader` / `<Entity>Writer` for persistence; role nouns for everything else (`KeyPairGenerator`, `TokenIssuer`, `TokenVerifier`, `PeerKeySet`, `PrivateKeySealer`, `JobLock`).
- Mocks are generated with mockery v3 (`make mocks`) into `port/mocks`; never hand-written.

## 5. Adapters

- Controllers parse the request, call one interactor, hand the result to their presenter. They never contain business rules.
- Presenters own every JSON shape and the error envelope `{"error":{"code","message","field"}}`. Public messages for 401 are always generic; the reason is logged.
- Repositories are the only importers of `internal/store`, map sqlc rows to domain types, translate `sql.ErrNoRows` into domain sentinels, and own transactions (`withTransaction`).
- Gateways are the only code that calls the peer over HTTP.

## 6. Infrastructure

- `infra/jwt` is the only package that knows the JWT library; `infra/jwks` the only one that parses JWKS documents; `infra/rsakey` the only one that touches RSA types.
- Frameworks and drivers are configured in `infra` and wired in `di`.

## 7. SQL and migrations

- Follow the naming standard in the root `CLAUDE.md`. Migrations are Up only (sqlc reads them as the schema), never edited once merged, no JSONB.
- Name every selected column; `SELECT *` is rejected by `sqlc vet` rules.

## 8. Tests

- Domain and infra: table tests with `testify`.
- Use cases: mockery mocks with `.EXPECT()`.
- Integration (`//go:build integration`): a fresh database per test, `di.NewWithDB` with a fake clock, the real HTTP surface through `httptest`.
