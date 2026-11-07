.DEFAULT_GOAL := help

COMPOSE ?= $(shell command -v podman >/dev/null 2>&1 && echo "podman compose" || echo "docker compose")
APPS := my_app partner_app

GITLEAKS := GOTOOLCHAIN=auto go run github.com/zricethezav/gitleaks/v8@v8.30.1

.PHONY: help env up dev db down down-v logs logs-my-app logs-partner-app ps test test-db lint generate docs vulncheck scan-deps lint-dockerfile scan-image secrets secrets-staged security hooks verify demo

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-17s %s\n", $$1, $$2}'

env: ## Create .env from .env.example when missing (fast rotation profile)
	@test -f .env || cp .env.example .env

up: env ## Build and start postgres, my_app and partner_app
	$(COMPOSE) up -d --build

dev: env ## Start postgres and both apps under air; Go changes rebuild and restart the app
	$(COMPOSE) -f docker-compose.yml -f docker-compose.dev.yml up -d --build

db: env ## Start only postgres (for host runs and integration tests)
	$(COMPOSE) up -d postgres

down: ## Stop and remove containers (database volume kept)
	$(COMPOSE) down

down-v: ## Stop and remove containers AND the database volume
	$(COMPOSE) down -v

logs: ## Follow the logs of both apps
	$(COMPOSE) logs -f my_app partner_app

logs-my-app: ## Follow the logs of my_app only
	$(COMPOSE) logs -f my_app

logs-partner-app: ## Follow the logs of partner_app only
	$(COMPOSE) logs -f partner_app

ps: ## Show service status
	$(COMPOSE) ps

test: ## Unit tests of both apps
	@for app in $(APPS); do echo "== $$app"; $(MAKE) -C $$app test || exit 1; done

test-db: ## Integration tests of both apps (needs `make db`)
	@for app in $(APPS); do echo "== $$app"; $(MAKE) -C $$app test-db || exit 1; done

lint: ## Lint both apps
	@for app in $(APPS); do echo "== $$app"; $(MAKE) -C $$app lint || exit 1; done

generate: ## Regenerate sqlc code, mocks and OpenAPI specs of both apps
	@for app in $(APPS); do $(MAKE) -C $$app generate || exit 1; done

docs: ## Regenerate the OpenAPI specs of both apps
	@for app in $(APPS); do $(MAKE) -C $$app docs || exit 1; done

vulncheck: ## Reachable known vulnerabilities in the Go dependencies of both apps (needs network)
	@for app in $(APPS); do echo "== $$app"; $(MAKE) -C $$app vulncheck || exit 1; done

scan-deps: ## Every known vulnerability in go.mod of both apps (needs network)
	@for app in $(APPS); do echo "== $$app"; $(MAKE) -C $$app scan-deps || exit 1; done

lint-dockerfile: ## Lint the Dockerfiles of both apps with hadolint
	@for app in $(APPS); do echo "== $$app"; $(MAKE) -C $$app lint-dockerfile || exit 1; done

scan-image: ## Build the runtime image of both apps and scan it with trivy (needs network)
	@for app in $(APPS); do echo "== $$app"; $(MAKE) -C $$app scan-image || exit 1; done

secrets: ## Scan the whole git history for leaked secrets; reviewed false positives live in .gitleaksignore
	$(GITLEAKS) git --redact --no-banner --gitleaks-ignore-path .gitleaksignore .

secrets-staged: ## Scan staged changes for secrets (run by the pre-commit hook)
	$(GITLEAKS) git --staged --redact --no-banner --gitleaks-ignore-path .gitleaksignore .

security: vulncheck scan-deps secrets lint-dockerfile scan-image ## Every security scan

hooks: ## Use the versioned git hooks in .githooks (gitleaks on every commit)
	git config core.hooksPath .githooks

verify: lint test test-db secrets ## Everything that gates a commit

demo: ## Send traffic through my_app for DURATION seconds and show which kid signed each request
	@./scripts/demo.sh $${DURATION:-300}
