# Gigmann Executive Cockpit — developer tasks
# Coverage is measured over ./internal/... (the meaningful, testable code).
# cmd/ (wiring) and generated code are excluded from the gate by design.

COVERAGE_THRESHOLD ?= 80
BACKEND_DIR := backend
FRONTEND_DIR := frontend

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

## ---- Backend ----
.PHONY: backend-build
backend-build: ## Build the Go API binary
	cd $(BACKEND_DIR) && go build -o bin/api ./cmd/api

.PHONY: backend-run
backend-run: ## Run the Go API locally
	cd $(BACKEND_DIR) && go run ./cmd/api

.PHONY: backend-test
backend-test: ## Run Go tests with race detector + coverage
	cd $(BACKEND_DIR) && go test -race -covermode=atomic -coverprofile=coverage.out ./internal/...
	cd $(BACKEND_DIR) && go tool cover -func=coverage.out | tail -n 1

.PHONY: backend-cover-gate
backend-cover-gate: backend-test ## Fail if backend coverage < $(COVERAGE_THRESHOLD)%
	@cd $(BACKEND_DIR) && total=$$(go tool cover -func=coverage.out | awk '/^total:/ {gsub("%","",$$3); print $$3}'); \
	echo "Backend coverage: $$total% (threshold $(COVERAGE_THRESHOLD)%)"; \
	awk -v t=$$total -v thr=$(COVERAGE_THRESHOLD) 'BEGIN{ exit !(t+0 >= thr+0) }' || \
	{ echo "FAIL: coverage $$total% < $(COVERAGE_THRESHOLD)%"; exit 1; }

.PHONY: backend-lint
backend-lint: ## Run golangci-lint
	cd $(BACKEND_DIR) && golangci-lint run ./...

.PHONY: backend-tidy
backend-tidy: ## go mod tidy
	cd $(BACKEND_DIR) && go mod tidy

## ---- Frontend ----
.PHONY: frontend-install
frontend-install: ## Install frontend deps
	cd $(FRONTEND_DIR) && npm install

.PHONY: frontend-test
frontend-test: ## Run frontend tests with coverage
	cd $(FRONTEND_DIR) && npm run test:coverage

.PHONY: frontend-lint
frontend-lint: ## Lint + typecheck frontend
	cd $(FRONTEND_DIR) && npm run lint && npm run typecheck

## ---- Local infra ----
.PHONY: dev-up
dev-up: ## Start local Postgres 16 + pgvector and Redis
	docker compose up -d

.PHONY: dev-down
dev-down: ## Stop local infra
	docker compose down

## ---- Aggregate ----
.PHONY: test
test: backend-cover-gate ## Run all gated tests (backend coverage gate)

.PHONY: lint
lint: backend-lint ## Run all linters
