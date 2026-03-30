# Identity Service

A Go-based authentication and **identity service** with JWT, refresh token rotation, role-based access control, token revocation, and rate limiting.

## 🔹 Features
<ul style="margin:0; padding-left:20px; line-height:1.2">
  <li><b>JWT-based authentication</b> (access & refresh tokens)</li>
  <li><b>Secure refresh token rotation</b></li>
  <li><b>Refresh token revocation</b> endpoint</li>
  <li><b>Role-based access control</b> (user/admin)</li>
  <li><b>Rate limiting</b> with Redis</li>
  <li><b>Middleware:</b>
    <ul style="margin:0; padding-left:15px; line-height:1.2">
      <li>Authentication</li>
      <li>Request context</li>
      <li>Structured logging</li>
      <li>Rate limiting</li>
      <li>Panic recovery</li>
      <li>Request ID tracking</li>
      <li>Role-based authorization</li>
    </ul>
  </li>
  <li><b>Graceful shutdown</b></li>
  <li><b>Centralized error handling</b></li>
  <li><b>Clean architecture</b> with a separate service layer</li>
  <li><b>Migration DB tables</b></li>
</ul>

## 🔹 Tech Stack
![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white) ![PostgreSQL](https://img.shields.io/badge/PostgreSQL-336791?style=for-the-badge&logo=postgresql&logoColor=white) ![Redis](https://img.shields.io/badge/Redis-DC382D?style=for-the-badge&logo=redis&logoColor=white) ![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)

## 🔹 Installation
```bash
# Cloning the repository
git clone https://github.com/Woun1zoN/go-identity-service.git
cd go-identity-service

# Create your own environment file
cp .env.example .env
# Edit .env to reflect your settings

# Install dependencies
go mod tidy

# Build and start all containers
docker-compose up --build
```

## 🔹 Usage
### Startup Logs
```bash
app-1    | 2023/12/01 12:00:00 Connected to DB
app-1    | 2023/12/01 12:00:00 Server started on http://localhost:8080
```
---
### CURL Requests Example ([Full API Documentation](https://github.com/Woun1zoN/go-identity-service/blob/main/documentation/api.md))
#### 🟡 POST `/register`

##### CURL Request:
```bash
curl -X POST http://localhost:8080/register \
-H "Content-Type: application/json" \
-d '{
  "email": "test@example.com",
  "password": "supersecret"
}'
```
##### Response:
```json
{
  "id": 1,
  "email": "test@example.com",
  "role": "user"
}
```
#### 🟡 POST `/login`

##### CURL Request:
```bash
curl -X POST http://localhost:8080/login \
-H "Content-Type: application/json" \
-d '{
  "email": "test@example.com",
  "password": "supersecret"
}'
```
##### Response:
```json
{
  "access_token": "access_token",
  "refresh_token": "refresh_token"
}
```
---
## 🔹 Project Structure
```bash
go-identity-service
├── cmd
│   └── main.go                  # application entry point
├── internal
│   ├── auth
│   │   └── jwt.go               # JWT handling (generation, verification)
│   ├── db
│   │   ├── connection.go        # database connection
│   │   └── migrate.go           # database schema migrations
│   ├── error_handling
│   │   ├── error-handling.go    # centralized error handling
│   │   └── errors.go
│   ├── handlers
│   │   ├── auth-handler.go      # HTTP endpoints for authentication
│   │   └── user-handler.go      # HTTP endpoints for users
│   ├── middleware
│   │   ├── auth.go              # authorization verification
│   │   ├── context.go           # adding context to requests
│   │   ├── logger.go            # logging
│   │   ├── rate-limiting.go     # request rate limiting
│   │   ├── recovery.go          # panic recovery
│   │   ├── request-id.go        # generating a unique request ID
│   │   └── roles.go             # role-based access
│   ├── models
│   │   ├── db.go                # database models
│   │   └── http.go              # models for the API
│   ├── repository
│   │   ├── token.go             # working with tokens (storage, validation)
│   │   └── user.go              # working with users in the database
│   ├── server
│   │   └── server.go            # HTTP server configuration and startup
│   └── service
│       ├── auth-service.go      # authentication business logic
│       └── user-service.go      # business logic for working with users
├── go.mod                       # go dependencies and modules
└── go.sum
```

## 🔹 Configuration
### Database Tables:
#### Table `users`
```sql
    Column     |            Type             | Collation | Nullable |              Default
---------------+-----------------------------+-----------+----------+-----------------------------------
 id            | integer                     |           | not null | nextval('users_id_seq'::regclass)
 email         | text                        |           | not null |
 password_hash | text                        |           | not null |
 created_at    | timestamp without time zone |           |          | now()
 role          | character varying(20)       |           |          | 'user'::character varying
```
#### Table `refresh_tokens`
```sql
   Column   |            Type             | Collation | Nullable | Default
------------+-----------------------------+-----------+----------+---------
 id         | uuid                        |           | not null |
 user_id    | integer                     |           | not null |
 token_hash | text                        |           | not null |
 expires_at | timestamp without time zone |           | not null |
 revoked    | boolean                     |           | not null | false
 created_at | timestamp without time zone |           | not null | now()
```
> **Tables are [automatically created](internal/db/migrate.go) via `RunMigrations()` on service startup.**
---
### Environment Variables (`.env`):
```env
# PostgreSQL
DB_USER=your_db_user       # database username
DB_PASSWORD=your_db_pass   # password
DB_NAME=your_db_name       # database name

# JWT
JWT_SECRET=changeme        # token signing secret (should be strong)
```
---
### Redis & Rate Limiting
#### Redis Configuration (from [main.go](https://github.com/Woun1zoN/go-identity-service/blob/main/cmd/main.go))
```go
rdb := redis.NewClient(&redis.Options{
    Addr:     "redis:6379",
    Password: "",
    DB:       0,
})
```
#### Individual limits for endpoints:
| Endpoint         | Limit       | Notes                      |
| ---------------- | ----------- | -------------------------- |
| `/register`      | 3 / minute  | open endpoint              |
| `/login`         | 5 / minute  | open endpoint              |
| `/refresh`       | 10 / minute | open endpoint              |
| `/admin/promote` | 1 / minute  | requires auth + admin role |
---
## 🔹 License & Contacts
This project is licensed under the [**MIT License**](LICENSE) © 2026 Wᴏᴜɴ†ᴢᴏN メ
### Contact me:
[![Discord](https://img.shields.io/badge/Discord-5865F2?style=for-the-badge&logo=discord&logoColor=white)](https://discord.com/users/1351287706164006982) [![Telegram](https://img.shields.io/badge/Telegram-2CA5E0?style=for-the-badge&logo=telegram&logoColor=white)](https://t.me/WountzoN)
