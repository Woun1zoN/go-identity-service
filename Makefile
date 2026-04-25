.PHONY: help env deps db wait test test-integration test-unit test-all clean dev dev-down

DB_CONTAINER=service_test
DB_USER=testuser

help: ## show available commands
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

dev: ## start dev environment
	docker-compose up --build

dev-down: ## stop dev environment
	docker-compose down

env: ## create .env from example
	cp -n .env.example .env || echo ".env already exists"

deps: ## install dependencies
	go mod tidy

db: ## start test database
	docker-compose -f docker-compose.test.yml up -d

wait: ## wait until DB is ready
	@echo "Waiting for DB to be ready..."
	@until docker exec $(DB_CONTAINER) pg_isready -U $(DB_USER); do sleep 1; done

test: db wait ## run all tests
	go clean -testcache
	go test ./tests/auth/... ./tests/middleware/...
	go test ./tests/integration/... -count=1

test-integration: db wait ## run integration tests
	go clean -testcache
	go test ./tests/integration/... -count=1

test-unit: ## run unit tests
	go test ./tests/auth/... ./tests/middleware/...

test-all: db wait ## run all test suites
	go clean -testcache
	go test ./tests/auth/... ./tests/middleware/...
	go test ./tests/migrations/...
	go test ./tests/integration/... -count=1

clean: ## stop test environment
	docker-compose -f docker-compose.test.yml down