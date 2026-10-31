.DEFAULT_GOAL := help

COMPOSE ?= $(shell command -v podman >/dev/null 2>&1 && echo "podman compose" || echo "docker compose")
APPS := my_app partner_app

.PHONY: help env up dev db down down-v logs logs-my-app logs-partner-app ps test test-db lint generate docs verify demo

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

verify: lint test test-db ## Everything that gates a commit

demo: ## Send traffic through my_app for DURATION seconds and show which kid signed each request
	@./scripts/demo.sh $${DURATION:-300}
