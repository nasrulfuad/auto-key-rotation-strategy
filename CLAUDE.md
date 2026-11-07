# Project rules

1. **English only.** All comments, variable/function/struct/type names, and log messages must be English. Never Indonesian in code, even if the prompt is Indonesian.
2. **Confirm before coding.** Before implementing a coding task, give a rough effort estimate and ask whether unit tests should be implemented as part of the change. Skip only for trivial one-liners.
3. **Clean architecture always.** domain (innermost, no framework/transport/db deps) -> usecase (application rules) -> adapter/infra (controllers, presenters, repositories, gateways, framework glue). Domain must not import net/http, database/sql, or framework code. Use cases depend only on interfaces, never on an implementation: repository interfaces live in `internal/adapter/repository`, gateway interfaces in `internal/adapter/gateway`, infrastructure interfaces in their infra package. Use cases never import the driving side (`adapter/controller`, `adapter/presenter`, `adapter/apidoc`, `infra/echo`, `infra/scheduler`, `infra/config`, `di`), nor `repository/postgres` or `gateway/http`. depguard enforces this.
4. **No comments, except interface contracts and configuration.** Never write comments in code — not inline, not on functions, ordinary structs or their fields. Code must be self-explanatory through naming and structure alone. Two exceptions get Go doc comments that start with the name and keep a blank line between documented members:
   - Exported interface types and every method of those interfaces: state the contract a caller relies on (results, the domain errors returned, side effects such as transactions or caching). `// FindActive returns ...`
   - Configuration structs (`config.Config`, the unexported `options` struct of every package, `Audiences`, `SigningKeySchedule`) and every field: state what the value controls, its environment variable and default when it has one, and the constraint it must meet. `// TokenTTL is the lifetime ... TOKEN_TTL, default 5m.`

   Nothing else gets a comment.
5. **Naming standard.** Every new or renamed identifier in every layer follows the tables below. When existing code disagrees with these rules, the rules win: rename it as part of the change that touches it, never copy the old pattern.
6. **No JSONB in the database, ever.** No JSONB columns, no `::JSONB` casts, no `jsonb_*` functions in queries, not even to pass a list of IDs or batch rows. Pass lists as typed arrays (`= ANY(sqlc.arg(ids)::UUID[])`, `unnest(...)`). Existing JSONB use is removed when the code around it is touched.
7. **Interface and implementation files.** Interfaces never share a file with an implementation. The file that implements an interface carries the `impl` suffix after the interface file's name. Every implementation has a compile-time assertion (`var _ repository.SigningKeyReader = (*SigningKeyRepository)(nil)`); when that would make a package import its consumer, the assertion goes in a `<name>impl_test.go`. Generated code (sqlc `store/querier.go`) is exempt.
8. **Functional options for constructors.** Every constructor takes its required dependencies and values as positional parameters and everything that has a sensible default as trailing `opts ...Option`. Each package declares one `type Option func(*options)`, an unexported `options` struct holding the defaults, `With<Name>` setters that ignore a nil dependency, and `newOptions(opts)` that applies them; every constructor of that package accepts `opts ...Option`. Never pass `nil` for an optional dependency and never add an `Options` / `Overrides` struct parameter. A package whose constructors have nothing optional (`sealer`, `systemclock`, `usecase/peersigningkey`) declares no `Option`. Each package tests its defaults and setters in `option_internal_test.go`.
9. **Security gates stay green.** `make verify` runs govulncheck, osv-scanner and gitleaks; `.github/workflows/security.yml` also runs gosec (inside `make lint`), hadolint and trivy for both apps. Fix a finding by upgrading (the Go patch in `go.mod` and `go.work`, the dependency, the base image digest) or by changing the code. Ignore only a reviewed false positive, and only where its scanner looks: gosec in `exclusions.rules` of the app's `.golangci.yml` with `path`, `text: "G<id>:"` and a YAML comment stating why it cannot apply (configuration, so rule 4 allows it); gitleaks by fingerprint in `.gitleaksignore`; osv-scanner in `<app>/osv-scanner.toml` with `reason` and `ignoreUntil`. Never `#nosec` or `//nolint:gosec`. Every scanner, base image and GitHub Action is pinned by version plus digest or commit SHA; Dependabot bumps them. A Dockerfile keeps the production stage named `runtime`, running as UID `10001`, and builds with the exact Go version of the `go` directive in `go.mod` (`make check-go-version`, run by `lint`); a Dependabot bump of the `golang` image needs the matching `go.mod` and `go.work` bump in the same change.

## Naming standard

### General

| Topic | Rule | Example |
|-------|------|---------|
| Entity names | Write the full entity name in exported identifiers, routes, error codes and SQL. Never shorten to `Item`, `Contact`, `Tier`. `CogsTier` is allowed only when the interface, type or path already says `Source`. | `ProductItem`, `SupplierContact`, `SourceCogsTier` |
| One entity, one name | An entity keeps its table name in every layer. | table `account` -> `domain.Account`, `AccountController`, `/admin/accounts`, `account:read` |
| Abbreviations | Spell words out. Allowed: `ID`, `URL`, `HTTP`, `JSON`, `SQL`, `UUID`, `SKU`, `SRP`, `COGS`, `DB`, `UC` (use case variable suffix in `di` only). | `MinimumGuaranteeTransaction`, not `MgTransaction` |
| Initialisms | Go casing: all caps or all lower, never mixed. | `ID`, `sourceRowID`, not `Id` |
| Verbs | Functions and methods start with a verb. Use `Find` (one), `List` (paged), `ListAll` (unpaged), `Count`, `Create`, `Update`, `Delete`, `Lookup`, `Check`, `Get` (use case and handler for one resource). Never `Load`, `Fetch`, `Insert` outside SQL. | `ListTopProducts`, `FindIdentity` |
| Predicates | Boolean-returning methods end in `Exists`. | `AliasExists`, `PairExists` |
| Packages | Lower case, one word or concatenated words, singular, no `shared` / `common` / `util` / `helper`. | `productsupplier`, `aliasassignment` |
| Go files | Lower case, words concatenated, no underscore. Underscore only for `_test.go` and one test qualifier: `_internal_test.go`, `_sql_test.go`, `_contract_test.go`. | `productsupplier.go`, `formrequestsourcecogstier.go`, `account_sql_test.go` |
| Test functions | `Test<Subject><Behaviour>`, no underscores. | `TestFormRequestFindByIDLoadsPayloadLabels` |

### IDs inside a form request

| Concept | Name | Column |
|---------|------|--------|
| Master record ID | `<Entity>ID` | `source_id`, `product_item_id` |
| Line ID, stable across events | `LineID`, `<Entity>LineID` | `line_id`, `source_line_id` |
| Projection row primary key | `<Entity>RowID` | `id` |

### Domain

| Topic | Rule | Example |
|-------|------|---------|
| Enum constants | `<TypeName><Value>`, full type name. | `FormRequestStatusDraft`, `StepActionApprove`, `PermissionProductRead` |
| Sentinel errors | `Err` + error code in PascalCase. | `ErrWorkflowAlreadyActive` for `WORKFLOW_ALREADY_ACTIVE` |
| Error constructors | Named after the error kind, optional suffix `Coded` / `Caused` / `f`. | `NotFound`, `BadRequestCoded`, `Internalf` |
| Error codes | `<FULL_ENTITY>_<CONDITION>` in `UPPER_SNAKE`. Conditions: `_TAKEN` (exists in master), `_DUPLICATED` (repeated inside one request), `_VERSION_CONFLICT`, `_NOT_FOUND`, `_IN_USE`, `TOO_MANY_<ENTITIES>`. | `PRODUCT_ITEM_ALIAS_TAKEN`, `TOO_MANY_SOURCE_COGS_TIERS` |
| Permission codes | `resource:action`, resource in `snake_case`. | `account:read`, `publisher:lookup` |

### Use case

| Topic | Rule | Example |
|-------|------|---------|
| Driven interfaces | `<Entity>Reader`, `<Entity>Writer`; others are role nouns. Declared outside `usecase`, see the interface table below. | `repository.AccountReader`, `infra/jwt.TokenIssuer` |
| Use-case-to-use-case interfaces | Declared in the consuming use case package in its own `<role>.go`. | `subscription.TokenSigner` in `tokensigner.go` |
| Interface methods | Verb first, see General. Unpaged lists use the `ListAll` prefix, never an `All` suffix. | `ListAllByProduct`, `FindMasterSet` |
| Interface parameter structs | `<Entity><Operation>`, in the same file as the interface. | `AccountCreate`, `FormRequestSourceDelete` |
| Filters | `<Entity>Filter`, `<Entity>LookupFilter`. Free-text search field is always `Query`. | `SupplierFilter.Query` |
| Interactors | `<Verb><Object>Interactor`, constructor `New<Verb><Object>Interactor`, single method `Execute(ctx, in)`. One interactor per file; file name is the interactor name without `Interactor`, lower case. Exception: package `lookup` omits the verb and uses a plural object. | `ListTopProductsInteractor` in `listtopproducts.go`; `lookup.PublishersInteractor` |
| Input / output | `<Interactor name without Interactor>Input` / `...Output`. | `ListTopProductsInput` |
| Dependency fields | Field and constructor parameter are the interface name in camelCase. | `accountReader repository.AccountReader` |
| No stutter | A type must not repeat its package name. | `session.Issuer`, not `session.SessionIssuer` |

### Dependency injection (`internal/di`)

| Topic | Rule | Example |
|-------|------|---------|
| Container options | `di.With<Dependency>` for every dependency a test may replace; no `Overrides` struct. | `di.NewWithDB(cfg, db, di.WithClock(clock))` |
| Local dependency variables | Interface name in camelCase. | `refreshTokenWriter` |
| Use case variables | `<verb><Object>UC`. | `listAccountsUC`, `checkPublisherAliasUC` |

### Repository (`internal/adapter/repository/postgres`)

| Topic | Rule | Example |
|-------|------|---------|
| Types | `<Entity>Repository` + `New<Entity>Repository`, named after the interface it implements. Never `Store` (that is the sqlc package). | `DatabaseProbeRepository` |
| Exported methods | Identical to the interface. | — |
| Mappers | `to<Target>[From<Source>]`. | `toAccount`, `toSourcesFromListRow` |
| Helpers | Full entity name, no `src` / `dec` / `tier` / `item` / `contact` prefixes. | `sourceFindRow`, `cogsTierWriteRow` |

### Interfaces and implementations

| Kind | Interface file | Implementation file | Mocks |
|------|----------------|---------------------|-------|
| Repository | `internal/adapter/repository/<entity>repository.go`, package `repository`; `<Entity>Reader` and `<Entity>Writer` with their parameter structs | `internal/adapter/repository/postgres/<entity>repositoryimpl.go` | `repository/repositorymocks` |
| Other repository roles | `internal/adapter/repository/<role>repository.go` | `postgres/<role>repositoryimpl.go` | `repository/repositorymocks` |
| Gateway | `internal/adapter/gateway/<entity>gateway.go`, package `gateway` | `internal/adapter/gateway/http/<entity>gatewayimpl.go` | `gateway/gatewaymocks` |
| Infrastructure | `internal/infra/<package>/<role>.go` | `internal/infra/<package>/<role>impl.go` | `<package>/<package>mocks` |
| Use case to use case | `internal/usecase/<consumer>/<role>.go` | `internal/usecase/<provider>/<role>impl.go` | — |

| Topic | Rule | Example |
|-------|------|---------|
| One implementation per file | A file holds one implementing type and its constructor; split when two types implement two interfaces. | `joblockrepositoryimpl.go`, `databaseproberepositoryimpl.go` |
| Impl tests | Named after the implementation file. | `tokenverifierimpl_test.go` |
| Mocks | Generated by mockery v3 from the interfaces listed in `.mockery.yml` (`make mocks`), never hand-written. | `repositorymocks.NewMockSigningKeyReader(t)` |
| Package names | An interface name must not repeat its package name; rename the package instead. | `systemclock.Clock`, not `clock.Clock` |

### Doc comments (rule 4 exceptions)

| Topic | Rule | Example |
|-------|------|---------|
| Where | Exported interface types and each of their methods; configuration structs, including each package's `options`, and each of their fields. Nowhere else: not on implementations, functions, ordinary structs, constants or tests. | `repository.SigningKeyReader`, `config.Config`, `jwks.Options` |
| First word | The name of the documented type, method or field. | `// FindActive returns ...`, `// TokenTTL is ...` |
| Interface method content | The contract: what it returns, which domain error it returns and when, and side effects (transaction, cache, lock, idempotency). Never restate the signature. | `// Create stores notification and reports whether it was new.` |
| Configuration field content | What the value controls, then the environment variable and its default (or `required`), then the constraint it must meet. | `// TokenTTL is the lifetime of every token this app signs; it must not exceed KEY_INACTIVE_DURATION.` / `// TOKEN_TTL, default 5m.` |
| Accuracy | Written from the implementation and the config loader, and updated in the same change that alters the behaviour, default or constraint. | — |
| Layout | Full sentences, wrapped near 120 columns, one blank line between documented methods or fields. | — |

### Functional options

| Topic | Rule | Example |
|-------|------|---------|
| File | `option.go` in the package, holding `Option`, `options`, the `With<Name>` setters and `newOptions`. | `usecase/signingkey/option.go` |
| Setters | `With<Name>`, the name of the dependency or value; nil dependencies are ignored so the default stays. | `WithClock`, `WithLogger`, `WithHTTPClient`, `WithRefreshMinimumInterval` |
| Defaults | Set in `newOptions`; a default that needs a required argument takes it there. | `newOptions(db, opts)` defaults the querier to `store.New(db)` |
| Use case defaults | `WithClock` → `systemclock.New()`, `WithUUIDGenerator` → `systemclock.NewUUIDGenerator()`, `WithLogger` → `logging.NopLogger{}`. | `signingkey.NewEnsureSigningKeysInteractor(reader, writer, generator, schedule, algorithm, signingkey.WithClock(clock))` |
| Required values | Never options: interfaces without a default, configuration values without a default, audiences, providers. | `NewRotateSigningKeyInteractor(jobLock, reader, writer, generator, lockName, schedule, algorithm, ...)` |
| Tests | `option_internal_test.go` covers the defaults, every setter and the nil case. | `TestNewOptionsIgnoresNilDependencies` |

### Controller and presenter

| Topic | Rule | Example |
|-------|------|---------|
| Controllers | `<Entity>Controller`; nested form request resources get their own `FormRequest<Entity>Controller`. | `AccountController`, `FormRequestProductController` |
| Handlers | Verb first, full entity name, same verb as the interactor they call. | `Decide`, `ListSupplierContacts`, `LookupPublishers` |
| Controller fields | Interactor name without `Interactor`, camelCase. Own presenter is `presenter`; extra presenters are `<entity>Presenter`. | `listStaleFormRequests`, `permissionPresenter` |
| Request structs | `<verb><Object>Request`, unexported. | `createAccountRequest` |
| Presenters | `<Entity>Presenter`; method name equals the handler it serves; a single-resource response shared by several handlers is `Get<X>`. | `FormRequestPresenter.ListSources` |
| Response structs | `<object>Body`, unexported. Never `ResponseBody`. | `accountBody` |
| `*echo.Context` variable (echo v5) | Always `ectx`. | `func (c *AccountController) Get(ectx *echo.Context) error` |

### HTTP contract

| Topic | Rule | Example |
|-------|------|---------|
| Path segments | `kebab-case`, plural, full entity name. | `/product-suppliers`, `/form-requests/:id/supplier-contacts` |
| Nested path segments | When the parent segment already names the owner, drop that prefix from the child segment so the word is not repeated. Keep the full name when the parent is a different entity. Go identifiers still use the full entity name. | `/suppliers/:id/contacts`, `/products/:id/items`, `/sources/:id/cogs-tiers`; but `/product-suppliers/:id/product-items` |
| Actions | Verb sub-path with `POST`. | `/form-requests/:id/decide`, `/admin/accounts/:id/deactivate` |
| Path parameters | `snake_case`; `:id` for the top-level resource. | `:round_id`, `:source_row_id` |
| Query parameters | `snake_case`. Search is `q`. Never reuse an entity name as a parameter name. | `change_type`, `data_source`, `group_code` |
| JSON fields | `snake_case`, same word as the database column. | `expires_at`, `cogs_value` |

### SQL queries (`query/*.sql`)

| Topic | Rule | Example |
|-------|------|---------|
| Files | One file per entity, named like the Go files. | `account.sql`, `refreshtoken.sql`, `formrequestproduct.sql` |
| Query names | `Insert`, `Update`, `Upsert`, `Delete`, `Find<X>By<Key>`, `List<X>` (paged, no `Paged` suffix), `ListAll<X>By<Key>` (unpaged), `Count<X>`, `Lookup<X>`, `<X>Exists`, `Lock<X>`. Always `By`, never `Of`. | `InsertAccount`, `FindRefreshTokenByHash`, `ListAllSourcesByProductSupplier` |
| List and count pairs | A `List<X>` query and its `Count<X>` keep identical `WHERE` clauses. | `ListSuppliers` / `CountSuppliers` |
| Comments | Only the `-- name:` line. | — |

### Database schema (`migrations/`)

| Topic | Rule | Example |
|-------|------|---------|
| Migration files | `NNNNN_snake_case_description.sql`, forward only, no Down section. The baseline is `00001_init.sql`. Never edit a merged migration; add a new one. | `00002_add_price_table.sql` |
| Migration content | Plain SQL statements only: no `DO` blocks, no procedural code, no comments except goose directives. Constraints are declared inline in `CREATE TABLE` with an explicit `CONSTRAINT <name>`; indexes follow their table. | — |
| Tables | Singular, `snake_case`. Child tables are prefixed with the parent. | `supplier_contact`, `form_request_source_payload_delta` |
| Columns | `snake_case`, full words. | `minimum_guarantee_transaction` |
| Primary key | `id`. | — |
| Foreign key column | `<referenced_table>_id`. | `product_supplier_id` |
| Timestamps | `TIMESTAMPTZ`, `<verb>_at`. Ranges use `<adjective>_from` / `<adjective>_until`. Never `_date` on a timestamp. | `expires_at`, `terminated_at`, `effective_from` |
| Booleans | `is_<adjective>`. | `is_active` |
| Ordering | `position` = display order, `step_order` = workflow step number, `sequence_number` = system-assigned running number. No other names. | — |
| Before values | Same column name with suffix `_before`. | `cogs_value_before` |
| Index | `<table>_<columns>_idx`, columns in index order. Partial or expression indexes add a qualifier before the columns. | `supplier_contact_supplier_id_name_idx`, `form_request_live_status_idx` |
| Unique constraint or index | `<table>_<columns>_key`. | `source_product_supplier_alias_key` |
| Foreign key constraint | `<table>_<column>_fk`. | `workflow_step_role_role_id_fk` |
| Check constraint | `<table>_<rule>_chk`. | `source_cogs_value_chk` |
| Primary key constraint | PostgreSQL default `<table>_pkey`. | — |
| Constraint names | Always explicit, always the full table name, at most 63 characters. Shorten the rule part, never the table name. | — |
| When to index | Every foreign key column, every column used in `WHERE` or `ORDER BY` of a list query, and every uniqueness rule the application relies on. | — |
| Generated types | A table whose name sqlc would singularize wrongly goes into `inflection_exclude_table_names` in `sqlc.yaml`. | `...PayloadDelta`, not `...PayloadDeltum` |
