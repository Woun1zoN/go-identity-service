DB_CONTAINER_NAME := service_test
DB_PASSWORD := 1234
DB_PORT := 5432

.PHONY: all env deps db up test clean

all: env deps db up test

env:
	cp -n .env.example .env || echo ".env already exists"

deps:
	go mod tidy

up:
	docker-compose up --build -d

test:
	go test ./tests/...

clean:
	docker rm -f $(DB_CONTAINER_NAME) || echo "Container not found"