package repository

import (
	"context"
	"time"

	"github.com/Woun1zoN/go-identity-service/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	DB *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) GetRefreshTokenByID(ctx context.Context, jti string) (*models.RefreshToken, error) {
	var token models.RefreshToken

	err := r.DB.QueryRow(ctx, "SELECT token_hash, revoked, expires_at FROM refresh_tokens WHERE id=$1", jti).Scan(&token.TokenHash, &token.Revoked, &token.ExpiresAt)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *UserRepository) RevokeRefreshToken(ctx context.Context, jti string) error {
	_, err := r.DB.Exec(ctx, "UPDATE refresh_tokens SET revoked = true WHERE id = $1", jti)
	return err
}

func (r *UserRepository) InsertRefreshToken(ctx context.Context, jti, userID, hash string, expiresAt time.Time) error {
	_, err := r.DB.Exec(ctx, "INSERT INTO refresh_tokens (id, user_id, token_hash, revoked, expires_at) VALUES ($1, $2, $3, false, $4) ", jti, userID, hash, expiresAt)
	return err
}