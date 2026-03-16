package service

import (
	"context"
	"errors"
	"time"
	"strconv"

	"github.com/Woun1zoN/go-identity-service/internal/repository"
	"github.com/Woun1zoN/go-identity-service/internal/models"
	"github.com/Woun1zoN/go-identity-service/internal/auth"

	"golang.org/x/crypto/bcrypt"

	"github.com/jackc/pgx/v5"
)

type Service struct {
	UserRepo *repository.UserRepository
}

func NewService(userRepo *repository.UserRepository) *Service {
	return &Service{
		UserRepo: userRepo,
	}
}

func (s *Service) RegisterUser(ctx context.Context, email, password string) (*models.UserResponse, error) {
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	userID, err := s.UserRepo.CreateUser(ctx, email, string(passHash))
	if err != nil {
		return nil, err
	}

	return &models.UserResponse{
		ID:    userID,
		Email: email,
	}, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*models.TokenResponse, error) {
	user, err := s.UserRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("unauthorized")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("unauthorized")
	}

	accessToken, err := auth.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshID, refreshHash, err := auth.GenerateRefreshToken(strconv.Itoa(user.ID))
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err := s.UserRepo.InsertRefreshToken(ctx, refreshID, strconv.Itoa(user.ID), refreshHash, expiresAt); err != nil {
		return nil, err
	}

	return &models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}