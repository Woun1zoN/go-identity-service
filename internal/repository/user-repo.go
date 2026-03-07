package repository

import (
	"context"
	"time"

	"github.com/Woun1zoN/go-identity-service/internal/models"
)

func (r *UserRepository) GetUserByID(ctx context.Context, userID int) (*models.UserResponse, error) {
	var response models.UserResponse
	var createdAt time.Time

	err := r.DB.QueryRow(ctx, "SELECT id, email, created_at FROM users WHERE id=$1", userID).Scan(&response.ID, &response.Email, &createdAt)
	if err != nil {
		return nil, err
	}

	response.Time = createdAt.Format("2006-01-02 15:04:05")
	return &response, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, email, passwordHash string) (int, error) {
	var userID int
	err := r.DB.QueryRow(ctx, "INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id",
		email, passwordHash).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.DB.QueryRow(ctx, "SELECT id, password_hash FROM users WHERE email = $1",
		email).Scan(&user.ID, &user.PasswordHash)
	if err != nil {
		return nil, err
	}
	return &user, nil
}