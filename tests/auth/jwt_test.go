package auth_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/Woun1zoN/go-identity-service/internal/auth"
	"github.com/Woun1zoN/go-identity-service/internal/models"
)

func TestGenerateAccessToken(t *testing.T) {
	jwtKey := []byte("test-jwt")
	authConfig := &auth.AuthConfig{
		JWTKey: jwtKey,
	}

	user := &models.User{
		ID: 123,
		Role: "admin",
	}

	tokenStr, err := authConfig.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			t.Fatalf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})

	if err != nil {
        t.Fatalf("failed to parse token: %v", err)
    }

    if !token.Valid {
        t.Fatal("token is invalid")
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        t.Fatal("claims are not of type MapClaims")
    }

    uidFloat, ok := claims["user_id"].(float64)
    if !ok {
        t.Fatalf("user_id claim is not a number: %v", claims["user_id"])
    }

    if int(uidFloat) != user.ID {
        t.Errorf("expected user_id %v, got %v", user.ID, int(uidFloat))
    }

    if claims["role"] != user.Role {
        t.Errorf("expected role %v, got %v", user.Role, claims["role"])
    }

    if claims["iss"] != "go-identity-service" {
        t.Errorf("expected issuer %v, got %v", "go-identity-service", claims["iss"])
    }

    if claims["aud"] != "go-api-users" {
        t.Errorf("expected audience %v, got %v", "go-api-users", claims["aud"])
    }

    exp := int64(claims["exp"].(float64))
    iat := int64(claims["iat"].(float64))
    now := time.Now().Unix()

    if exp <= now {
        t.Errorf("exp is in the past: %v", exp)
    }

    if iat > now {
        t.Errorf("iat is in the future: %v", iat)
    }
}

func TestGenerateRefreshToken(t *testing.T) {
	jwtKey := []byte("test-jwt")
	authConfig := &auth.AuthConfig{
		JWTKey: jwtKey,
	}

	userID := "123"

	tokenStr, refreshID, hashToken, err := authConfig.GenerateRefreshToken(userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			t.Fatalf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})
	if err != nil {
        t.Fatalf("failed to parse token: %v", err)
    }

	if !token.Valid {
        t.Fatal("token is invalid")
    }

	claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        t.Fatal("claims are not of type MapClaims")
    }

	if claims["user_id"] != userID {
        t.Errorf("expected user_id %v, got %v", userID, claims["user_id"])
    }

    if claims["jti"] != refreshID {
        t.Errorf("expected jti %v, got %v", refreshID, claims["jti"])
    }

    if claims["iss"] != "go-identity-service" {
        t.Errorf("expected issuer %v, got %v", "go-identity-service", claims["iss"])
    }

    if claims["aud"] != "go-api-users" {
        t.Errorf("expected audience %v, got %v", "go-api-users", claims["aud"])
    }

	exp := int64(claims["exp"].(float64))
    iat := int64(claims["iat"].(float64))
    now := time.Now().Unix()

	if exp <= now {
        t.Errorf("exp is in the past: %v", exp)
    }

    if iat > now {
        t.Errorf("iat is in the future: %v", iat)
    }

	expectedHash := auth.HashToken(tokenStr)
	if hashToken != expectedHash {
        t.Errorf("hash mismatch: expected %v, got %v", expectedHash, hashToken)
    }
}