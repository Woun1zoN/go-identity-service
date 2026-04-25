package service

import (
	"context"
	"fmt"
	"errors"
	"time"
	"strconv"

	"github.com/Woun1zoN/go-identity-service/internal/models"
	"github.com/Woun1zoN/go-identity-service/internal/auth"
	"github.com/Woun1zoN/go-identity-service/internal/repository"

	"github.com/golang-jwt/jwt/v5"
)


var ErrInvalidToken = errors.New("invalid token")


func (s *Service) GetUserByID(ctx context.Context, id int) (*models.UserResponse, error) {
	return s.UserRepo.GetUserByID(ctx, id)
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*models.TokenResponse, error) {
	token, err := jwt.Parse(refreshToken, func(t *jwt.Token) (interface{}, error) {
    if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("unexpected signing method")
    }
    return s.Auth.JWTKey, nil
    })
    if err != nil {
		return nil, ErrInvalidToken
	}

	if !token.Valid {
        return nil, ErrInvalidToken
    }

	claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return nil, ErrInvalidToken
	}

    jti, ok := claims["jti"].(string)
	if !ok {
        return nil, ErrInvalidToken
    }
    userIDstr, ok := claims["user_id"].(string)
	if !ok {
        return nil, ErrInvalidToken
    }

	iss, ok := claims["iss"].(string)
	if !ok {
        return nil, ErrInvalidToken
    }
    aud, ok := claims["aud"].(string)
	if !ok {
        return nil, ErrInvalidToken
    }

	if iss != "go-identity-service" || aud != "go-api-users" {
        return nil, ErrInvalidToken
    }

	hash := auth.HashToken(refreshToken)

	storedToken, err := s.UserRepo.GetRefreshTokenByID(ctx, jti)
    if err != nil {
        if errors.Is(err, repository.ErrNotFound) {
            return nil, ErrInvalidToken
        }
    return nil, err
    }

	if storedToken.TokenHash != hash || storedToken.Revoked || time.Now().After(storedToken.ExpiresAt) {
        return nil, ErrInvalidToken
    }

	if err := s.UserRepo.RevokeRefreshToken(ctx, jti); err != nil {
		return nil, err
	}

	userID, err := strconv.Atoi(userIDstr)
    if err != nil {
		return nil, ErrInvalidToken
	}

	user, err := s.UserRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	tokenUser := &models.User{
		ID:   user.ID,
		Role: user.Role,
	}

	access, err := s.Auth.GenerateAccessToken(tokenUser)
	if err != nil {
		return nil, err
	}

	refresh, newJTI, _, err := s.Auth.GenerateRefreshToken(strconv.Itoa(user.ID))
	if err != nil {
		return nil, err
	}

	newHash := auth.HashToken(refresh)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	if err := s.UserRepo.InsertRefreshToken(ctx, newJTI, strconv.Itoa(user.ID), newHash, expiresAt); err != nil {
		return nil, err
	}

	return &models.TokenResponse{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return errors.New("empty token")
	}

	token, err := jwt.Parse(refreshToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.Auth.JWTKey, nil
	})
	if err != nil {
		return err
	}

	if !token.Valid {
		return errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("invalid claims")
	}

	iss, _ := claims["iss"].(string)
	aud, _ := claims["aud"].(string)
	if iss != "go-identity-service" || aud != "go-api-users" {
		return errors.New("invalid issuer")
	}

	jti, ok := claims["jti"].(string)
	if !ok {
		return errors.New("jti missing")
	}

	return s.UserRepo.RevokeRefreshToken(ctx, jti)
}

func (s *Service) PromoteUser(ctx context.Context, userID int) error {
	return s.UserRepo.SetRole(ctx, userID, "admin")
}