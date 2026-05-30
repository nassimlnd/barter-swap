.DEFAULT_GOAL := help

COMPOSE       := docker compose
COMPOSE_TEST  := docker compose -f docker-compose.yml -f docker-compose.test.yml

.PHONY: help up down logs sh psql rebuild ps test test-local lint fmt vet tidy seed cover clean

help: ## Show this help
	@awk 'BEGIN{FS=":.*##"; printf "Targets:\n"} /^[a-zA-Z_-]+:.*##/{printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

up: ## Build + start db & api (dev mode with live reload)
	$(COMPOSE) up -d --build
	@echo "API   -> http://localhost:8080"
	@echo "DB    -> postgres://barter:barter@localhost:5432/barterswap"

down: ## Stop and remove containers + volumes
	$(COMPOSE) down -v

logs: ## Tail api logs
	$(COMPOSE) logs -f api

ps: ## List running containers
	$(COMPOSE) ps

sh: ## Shell into api container
	$(COMPOSE) exec api sh

psql: ## psql shell into the db container
	$(COMPOSE) exec db psql -U barter -d barterswap

rebuild: ## Full rebuild (down + up)
	$(COMPOSE) down
	$(COMPOSE) up -d --build

test: ## Run tests inside an isolated test container (Postgres tmpfs)
	$(COMPOSE_TEST) run --rm test
	$(COMPOSE_TEST) down -v

test-local: ## Run tests against the running dev DB (TEST_DATABASE_URL set)
	TEST_DATABASE_URL=postgres://barter:barter@localhost:5432/barterswap?sslmode=disable \
		go test -v -cover ./...

cover: ## Generate HTML coverage report (requires test-local first)
	go tool cover -html=coverage.out -o coverage.html
	@echo "Open coverage.html in your browser"

lint: ## gofmt + go vet inside container
	$(COMPOSE) exec api sh -c "gofmt -l . && go vet ./..."

fmt: ## Run gofmt on host
	gofmt -w .

vet: ## Run go vet on host
	go vet ./...

tidy: ## Run go mod tidy (host)
	go mod tidy

seed: ## Trigger /debug/seed (requires LOG_LEVEL=debug or a seed flag)
	curl -X POST http://localhost:8080/debug/seed

clean: ## Remove build artifacts
	rm -rf tmp coverage.out coverage.html
