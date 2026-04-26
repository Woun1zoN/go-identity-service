package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	"github.com/Woun1zoN/go-identity-service/internal/auth"
	"github.com/Woun1zoN/go-identity-service/internal/models"
	"github.com/Woun1zoN/go-identity-service/internal/repository"
)

func ParseJWT(t *testing.T, tokenStr string) jwt.MapClaims {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		require.Equal(t, jwt.SigningMethodHS256, token.Method)
		return []byte("testkey"), nil
	})
	require.NoError(t, err)
	require.True(t, token.Valid)

	claims, ok := token.Claims.(jwt.MapClaims)
	require.True(t, ok)

	return claims
}

func TestLoginUser(t *testing.T) {
	CleanDB(t)

	email := "test@example.com"
	password := "supersecret"

	hash, _ := auth.HashPassword(password)

	repo := repository.NewUserRepository(testDB)

	userID, _, err := repo.CreateUser(context.Background(), email, hash)
	require.NoError(t, err)

	body := fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, password)
	r := httptest.NewRequest("POST", "/login", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	TestRouter.ServeHTTP(w, r)

	require.Equal(t, http.StatusOK, w.Code)

	var resp models.TokenResponse
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	require.NotEmpty(t, resp.AccessToken)
	require.NotEmpty(t, resp.RefreshToken)

	accessClaims := ParseJWT(t, resp.AccessToken)
	refreshClaims := ParseJWT(t, resp.RefreshToken)

	accessUserID, ok := accessClaims["user_id"].(float64)
	require.True(t, ok)
	require.Equal(t, float64(userID), accessUserID)

	refreshUserID, ok := refreshClaims["user_id"].(string)
	require.True(t, ok)
	require.Equal(t, strconv.Itoa(userID), refreshUserID)
}