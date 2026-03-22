package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"time"
	"fmt"

	"github.com/Woun1zoN/go-identity-service/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthConfig struct {
    JWTKey []byte
}

func (auth *AuthConfig) GetJWTKey() error {
    key := os.Getenv("JWT_SECRET")
    if key == "" {
        return fmt.Errorf("There are no secrets!")
    }
    auth.JWTKey = []byte(key)
	return nil
}

func (auth *AuthConfig) GenerateAccessToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"iss":     "go-identity-service",
        "aud":     "go-api-users",
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(auth.JWTKey)
}

func (auth *AuthConfig) GenerateRefreshToken(userID string) (string, string, string, error) {
	refreshID := uuid.NewString()

	claims := jwt.MapClaims{
		"user_id": userID,
		"jti":     refreshID,
		"iss":     "go-identity-service",
		"aud":     "go-api-users", 
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	signing := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	token, err := signing.SignedString(auth.JWTKey)
	if err != nil {
		return "", "", "", err
	}

	hash_token := HashToken(token)

	return token, refreshID, hash_token, nil
}

func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}