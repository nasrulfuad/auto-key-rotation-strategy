.DEFAULT_GOAL := help

TEST_DATABASE_ADMIN_URL ?= postgres://jwks:jwks@localhost:55432/postgres

CONTAINER ?= $(shell command -v podman >/dev/null 2>&1 && echo podman || echo docker)

GOVULNCHECK    := go run golang.org/x/vuln/cmd/govulncheck@v1.9.0
OSV_SCANNER    := GOTOOLCHAIN=auto go run github.com/google/osv-scanner/v2/cmd/osv-scanner@v2.7.0
TRIVY_IMAGE    := docker.io/aquasec/trivy:0.76.2@sha256:8f6bb947dbc528902e11f2f5be52f7b519f6a39d5d148f41166f2040499768fb
HADOLINT_IMAGE := docker.io/hadolint/hadolint:v2.15.1@sha256:32dac94127fd60b7b7e3fbfc65e1383b9b5e25c9bfd7b8536de7a539fe68a12d
SCAN_IMAGE     := localhost/partner-app:scan

.PHONY: help generate sqlc mocks mocks-check docs docs-check fmt check-go-version lint test test-db vulncheck scan-deps lint-dockerfile scan-image security verify run

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-16s %s\n", $$1, $$2}'

generate: sqlc mocks docs ## Regenerate internal/store, the mocks and the OpenAPI spec

sqlc: ## Regenerate internal/store from query/*.sql and migrations/
	sqlc generate

mocks: ## Regenerate the mocks of the interfaces listed in .mockery.yml (run after changing one of them)
	go tool mockery

mocks-check: mocks ## Fail when the committed mocks have drifted from the interfaces
	git diff --exit-code -- $$(find internal -type d -name '*mocks')

docs: ## Regenerate api/openapi.json from the swag annotations
	go generate .

docs-check: docs ## Fail when the committed OpenAPI spec has drifted from the annotations
	git diff --exit-code -- api

fmt: ## Format Go sources and tidy the module
	gofmt -w .
	go mod tidy

check-go-version: ## Fail when the Dockerfile builds with a different Go version than go.mod declares
	@want=$$(awk '/^go / {print $$2}' go.mod); \
	for got in $$(sed -n 's#^FROM docker.io/library/golang:\([0-9.]*\)-.*#\1#p' Dockerfile); do \
		test "$$got" = "$$want" || { echo "Dockerfile builds with Go $$got but go.mod declares go $$want"; exit 1; }; \
	done

lint: check-go-version ## Go version check, gofmt check, go vet, golangci-lint (including integration tests)
	@test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)
	go vet ./...
	golangci-lint run ./...
	golangci-lint run --build-tags integration ./internal/integration/

test: ## Tests that need no database
	go test ./...

test-db: ## Integration tests against Postgres (needs `make db` or `make up` at the repo root)
	TEST_DATABASE_ADMIN_URL=$(TEST_DATABASE_ADMIN_URL) go test -tags=integration -count=1 ./internal/integration/

vulncheck: ## Reachable known vulnerabilities in Go dependencies and the stdlib (needs network)
	$(GOVULNCHECK) ./...

scan-deps: ## Every known vulnerability in go.mod, reachable or not; ignores live in osv-scanner.toml (needs network)
	$(OSV_SCANNER) scan source --config osv-scanner.toml -L go.mod

lint-dockerfile: ## Lint the Dockerfile with hadolint (config in the repo root .hadolint.yaml)
	$(CONTAINER) run --rm -i -v "$(CURDIR)/../.hadolint.yaml:/hadolint.yaml:ro" $(HADOLINT_IMAGE) \
		hadolint --config /hadolint.yaml - < Dockerfile

scan-image: ## Build the runtime image and fail on fixable HIGH/CRITICAL vulnerabilities (needs network)
	$(CONTAINER) build --target runtime -t $(SCAN_IMAGE) -f Dockerfile .
	mkdir -p tmp
	rm -f tmp/scan-image.tar
	$(CONTAINER) save -o tmp/scan-image.tar $(SCAN_IMAGE)
	$(CONTAINER) run --rm -v "$(CURDIR)/tmp:/scan:ro" -v key-rotation-trivy-cache:/root/.cache/trivy $(TRIVY_IMAGE) \
		image --input /scan/scan-image.tar --severity HIGH,CRITICAL --ignore-unfixed --exit-code 1 --no-progress

security: vulncheck scan-deps lint-dockerfile scan-image ## Every security scan of this app

verify: lint test test-db vulncheck scan-deps lint-dockerfile scan-image ## Everything that gates a commit, same scans as CI

run: ## Run the server on the host (reads .env when present)
	go run ./cmd/server
