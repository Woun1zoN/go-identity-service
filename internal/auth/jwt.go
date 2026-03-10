package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"time"

	"github.com/Woun1zoN/go-identity-service/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var JwtKey = []byte(os.Getenv("JWT_SECRET"))

func GenerateAccessToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(JwtKey)
}

func GenerateRefreshToken(userID string) (string, string, string, error) {
	refreshID := uuid.NewString()

	claims := jwt.MapClaims{
		"user_id": userID,
		"jti":     refreshID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	singing := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	token, err := singing.SignedString(JwtKey)
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