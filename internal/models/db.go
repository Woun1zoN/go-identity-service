package models

import (
    "time"
)

type RefreshToken struct {
	TokenHash string
	Revoked   bool
	ExpiresAt time.Time
}

type User struct {
	ID           int
	PasswordHash string
}