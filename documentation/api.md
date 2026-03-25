# 🔸 API Documentation

## 🔻 Navigation
   * [**API Endpoint Overview**](#api-endpoint-overview)
   * [**API Usage (cURL)**](#api-usage-curl)
      * [**Endpoint `/register`**](#post-register)
      * [**Endpoint `/login`**](#post-login)
      * [**Endpoint `/profile`**](#get-profile)
      * [**Endpoint `/refresh`**](#post-refresh)
      * [**Endpoint `/logout`**](#post-logout)
      * [**Endpoint `/admin/promote`**](#post-admin-promote)

## 🔻 API Endpoint Overview
| Endpoint        | Method | Action                                      | Rate Limit               |
|-----------------|-------|-------------------------------------------------|-------------------------|
| `/register`     | POST  | Registers a user (email + password), hashes the password, and saves it to the database | 3 requests / minute     |
| `/login`        | POST  | Verifies email/password and issues access and refresh tokens | 5 requests / minute     |
| `/profile`      | GET   | Returns user data based on the userID from the JWT (via middleware) | —                       |
| `/refresh`      | POST  | Refreshes access/refresh tokens using the refresh token | 10 requests / minute    |
| `/logout`       | POST  | Revokes the refresh token             | —                       |
| `/admin/promote`| POST  | Promotes the user to the admin role              | 1 request / minute      |

## 🔻 API Usage (cURL)
### 🟡 POST `/register`

#### CURL Request:
```bash
curl -X POST http://localhost:8080/register \
-H "Content-Type: application/json" \
-d '{
  "email": "test@example.com",
  "password": "supersecret"
}'
```
#### Request Body
```json
{
  "email": "test@example.com",
  "password": "supersecret"
}
```
#### Response:
```json
{
  "id": 1,
  "email": "test@example.com",
  "role": "user"
}
```
---
### 🟡 POST `/login`

#### CURL Request:
```bash
curl -X POST http://localhost:8080/login \
-H "Content-Type: application/json" \
-d '{
  "email": "test@example.com",
  "password": "supersecret"
}'
```
#### Request Body
```json
{
  "email": "test@example.com",
  "password": "supersecret"
}
```
#### Response:
```json
{
  "access_token": "access_token",
  "refresh_token": "refresh_token"
}
```
---
### 🟢 GET `/profile`

#### Request:
```bash
curl -X GET "http://localhost:8080/profile" \
-H "Authorization: Bearer <YOUR_ACCESS_TOKEN>" \
-H "Content-Type: application/json"
```
#### Authorization Header:
```bash
Bearer <YOUR_ACCESS_TOKEN> # <access_token> from /login response
```
#### Response:
```json
{
  "id": 1,
  "email": "test@example.com",
  "role": "user",
  "time": "2026-03-24 20:00:00"
}
```
---
### 🟡 POST `/refresh`

#### Request:
```bash
curl -X POST "http://localhost:8080/refresh" \
-H "Content-Type: application/json" \
-d '{
    "refresh_token": "<YOUR_REFRESH_TOKEN>"
  }'
```
#### Request Body
```json
{
  "refresh_token": "eyJhbGci..."
}
```
#### Response:
```json
{
  "access_token": "new_access_token",
  "refresh_token": "new_refresh_token"
}
```
---
### 🟡 POST `/logout`

#### Request:
```bash
curl -X POST http://localhost:8080/logout \
-H "Content-Type: application/json" \
-d '{
  "refresh_token": "<YOUR_REFRESH_TOKEN>"
}'
```
#### Request Body
```json
{
  "refresh_token": "eyJhbGci..."
}
```
#### Response:
```json
{
  "message": "Successfully logged out"
}
```
---
### 🟡 POST `/admin/promote`

#### Request:
```bash
curl -X POST http://localhost:8080/admin/promote \
-H "Authorization: Bearer <YOUR_ACCESS_TOKEN>" \
-H "Content-Type: application/json" \
-d '{
  "user_id": 1
}'
```
#### Authorization Header:
```bash
Bearer <YOUR_ACCESS_TOKEN>
```
#### Request Body
```json
{
  "user_id": 1
}
```
#### Response:
```json
{
  "message": "User promoted"
}
```
