FROM golang:1.26-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o app ./cmd/main.go

FROM alpine:latest

RUN apk add --no-cache postgresql-client

WORKDIR /app

COPY --from=build /app/app .
COPY --from=build /app/internal/db/migrations ./internal/db/migrations

EXPOSE 8080

CMD ["sh", "-c", "until pg_isready -h db -p 5432; do echo waiting for db; sleep 2; done; ./app"]