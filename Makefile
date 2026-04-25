.PHONY: env deps db up wait test test-integration test-unit test-all clean dev dev-down

DB_CONTAINER=service_test
DB_USER=testuser

dev:
	docker-compose up --build

dev-down:
	docker-compose down

env:
	cp -n .env.example .env || echo ".env already exists"

deps:
	go mod tidy

db:
	docker-compose -f docker-compose.test.yml up -d

wait:
	@echo "Waiting for DB to be ready..."
	@until docker exec $(DB_CONTAINER) pg_isready -U $(DB_USER); do sleep 1; done

test: db wait
	go clean -testcache
	go test ./tests/auth/... ./tests/middleware/...
	go test ./tests/integration/... -count=1

test-integration: db wait
	go clean -testcache
	go test ./tests/integration/... -count=1

test-unit:
	go test ./tests/auth/... ./tests/middleware/...

test-all: db wait
	go clean -testcache
	go test ./tests/auth/... ./tests/middleware/...
	go test ./tests/migrations/...
	go test ./tests/integration/... -count=1

clean:
	docker-compose -f docker-compose.test.yml down