.DEFAULT_GOAL := help

TEST_DATABASE_ADMIN_URL ?= postgres://jwks:jwks@localhost:55432/postgres

.PHONY: help generate sqlc mocks mocks-check docs docs-check fmt lint test test-db verify run

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-14s %s\n", $$1, $$2}'

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

lint: ## gofmt check, go vet, golangci-lint (including integration tests)
	@test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)
	go vet ./...
	golangci-lint run ./...
	golangci-lint run --build-tags integration ./internal/integration/

test: ## Tests that need no database
	go test ./...

test-db: ## Integration tests against Postgres (needs `make db` or `make up` at the repo root)
	TEST_DATABASE_ADMIN_URL=$(TEST_DATABASE_ADMIN_URL) go test -tags=integration -count=1 ./internal/integration/

verify: lint test test-db ## Everything that gates a commit

run: ## Run the server on the host (reads .env when present)
	go run ./cmd/server
