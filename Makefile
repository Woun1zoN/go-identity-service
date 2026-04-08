.PHONY: env deps db up wait test clean

env:
	cp -n .env.example .env || echo ".env already exists"

deps:
	go mod tidy

db:
	docker-compose -f docker-compose.test.yml up -d

wait:
	@echo "Waiting for DB to be ready..."
	@until docker exec service_test pg_isready -U testuser; do sleep 1; done

test: db wait
	go test ./tests/...

clean:
	docker-compose -f docker-compose.test.yml down -v