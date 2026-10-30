# Clean Architecture Rules

These rules apply to every change in `my_app`. They mirror `pricing/backend/rules/rules.md`, reduced to what this service contains.

## 1. Rings

```
cmd/server          entry point
  internal/di       composition root, the only place that sees the whole graph
    internal/adapter    controller/http, presenter, apidoc
    internal/infra      config, echo, postgres, scheduler
    internal/usecase    interactors
    internal/adapter    repository, gateway                         (driven interfaces only)
      internal/adapter  repository/postgres, gateway/http           (implementations)
        internal/store  sqlc-generated code
    internal/infra      systemclock, logging, rsakey, sealer, jwt, jwks   (<role>.go interface + <role>impl.go implementation)
    internal/domain     entities, value objects, rules
```

Repository and gateway interfaces live in the layer packages `adapter/repository` (`<entity>repository.go`) and `adapter/gateway` (`<entity>gateway.go`); use cases import those, never `repository/postgres` or `gateway/http`. Infrastructure interfaces live in the implementing package, each in its own `<role>.go` file, with the implementation in `<role>impl.go` (`infra/jwt/tokenissuer.go` and `infra/jwt/tokenissuerimpl.go`). Use cases never import the driving side: `adapter/controller`, `adapter/presenter`, `adapter/apidoc`, `infra/echo`, `infra/scheduler`, `infra/config` and `di`. depguard in `.golangci.yml` fails the lint when they do.

## 2. Domain (`internal/domain`)

- Standard library and `github.com/google/uuid` only. `net/http`, `database/sql` and `crypto/rsa` are denied: key material is stored as bytes (`PublicKey`, `Secret`) and converted by `infra/rsakey`.
- Holds every rule that can be expressed without I/O: the rotation schedule (`SigningKeySchedule.PlanNext`), key status, token expectations, identifier validation, retry backoff.
- Functions that depend on time take `now` as an argument.

## 3. Use cases (`internal/usecase`)

- One interactor per file, `<Verb><Object>Interactor` with a single `Execute(ctx, in)`.
- Dependencies are interfaces (`repository.SigningKeyReader`, `gateway.PartnerEntitlementGateway`, `systemclock.Clock`, `logging.Logger`), injected through the constructor and named after the interface.
- `Clock.Now()` is called once per `Execute` and passed down.
- May not import Echo, `database/sql`, `golang-jwt`, `internal/store` or the driving adapters and infrastructure listed in section 1.
- A dependency implemented by another use case (the token `Signer`) is consumed through a small interface declared in the consuming package (`TokenSigner`).

## 4. Driven interfaces

- Repository: `internal/adapter/repository/<entity>repository.go`, package `repository`, holding `<Entity>Reader` / `<Entity>Writer` and their parameter structs (`SigningKeyCreate`, `SigningKeyRotate`). `JobLock` and `DatabaseProbe` have their own files. Implemented in `repository/postgres` with a compile-time assertion (`var _ repository.SigningKeyReader = (*SigningKeyRepository)(nil)`).
- Gateway: `internal/adapter/gateway/<entity>gateway.go`, package `gateway`, implemented in `gateway/http`.
- Infrastructure: one `<role>.go` per interface in the implementing package (`KeyPairGenerator`, `PrivateKeySealer`, `TokenIssuer`, `TokenVerifier`, `PeerKeySet`, `Clock`, `UUIDGenerator`, `Logger`); the implementation file carries the `impl` suffix.
- Interfaces never share a file with an implementation.
- Mocks are generated with mockery v3 (`make mocks`, interfaces listed in `.mockery.yml`) into a `<package>mocks` directory next to the interfaces (`repository/repositorymocks`, `gateway/gatewaymocks`, `systemclock/systemclockmocks`); never hand-written.

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
