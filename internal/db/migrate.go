package db

import (
	"context"
	"fmt"
)

func (s *DBServer) RunMigrations(ctx context.Context) error {
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		role VARCHAR(20) DEFAULT 'user'
	);`

	_, err := s.DB.Exec(ctx, usersTable)
	if err != nil {
		return fmt.Errorf("error creating users table: %w", err)
	}

	tokensTable := `
	CREATE TABLE IF NOT EXISTS refresh_tokens (
		id UUID PRIMARY KEY,
		user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		token_hash TEXT NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		revoked BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = s.DB.Exec(ctx, tokensTable)
	if err != nil {
		return fmt.Errorf("error creating refresh_tokens table: %w", err)
	}

	return nil
}