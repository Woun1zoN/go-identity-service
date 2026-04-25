package integration_test

import (
	"testing"
	"context"
	"strconv"
	"time"
	"fmt"
	"net/http/httptest"
	"net/http"
	"strings"
	"encoding/json"

	"github.com/stretchr/testify/require"

	"github.com/Woun1zoN/go-identity-service/internal/repository"
	"github.com/Woun1zoN/go-identity-service/internal/auth"
	"github.com/Woun1zoN/go-identity-service/internal/models"
)

func TestRefreshToken(t *testing.T) {
	CleanDB(t)

	ctx := context.Background()
	repo := repository.NewUserRepository(testDB)

	email := "test@example.com"
    password := "supersecret"

	hash, _ := auth.HashPassword(password)
	userID, err := repo.CreateUser(ctx, email, hash)
	require.NoError(t, err)

	authService := &auth.AuthConfig{
		JWTKey: []byte("testkey"),
	}

	refreshToken, jti, _, err := authService.GenerateRefreshToken(strconv.Itoa(userID))
	require.NoError(t, err)

	tokenHash := auth.HashToken(refreshToken)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	err = repo.InsertRefreshToken(ctx, jti, strconv.Itoa(userID), tokenHash, expiresAt)
	require.NoError(t, err)

	body := fmt.Sprintf(`{"refresh_token":"%s"}`, refreshToken)

	r := httptest.NewRequest("POST", "/refresh", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	TestRouter.ServeHTTP(w, r)

	require.Equal(t, http.StatusOK, w.Code)

	var response models.TokenResponse
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	require.NotEmpty(t, response.AccessToken)
    require.NotEmpty(t, response.RefreshToken)

	var revoked bool
	err = testDB.QueryRow(ctx,
        "SELECT revoked FROM refresh_tokens WHERE id=$1", jti,
    ).Scan(&revoked)

	require.NoError(t, err)
    require.True(t, revoked)
}

func TestRefresh_InvalidToken(t *testing.T) {
    CleanDB(t)

    body := `{"refresh_token":"invalid.token"}`

    r := httptest.NewRequest("POST", "/refresh", strings.NewReader(body))
    r.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()
    TestRouter.ServeHTTP(w, r)

    require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRefresh_NotInDB(t *testing.T) {
    CleanDB(t)

    authService := &auth.AuthConfig{
        JWTKey: []byte("testkey"),
    }

    refreshToken, _, _, err := authService.GenerateRefreshToken("123")
    require.NoError(t, err)

    body := fmt.Sprintf(`{"refresh_token":"%s"}`, refreshToken)

    r := httptest.NewRequest("POST", "/refresh", strings.NewReader(body))
    r.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()
    TestRouter.ServeHTTP(w, r)

    require.Equal(t, http.StatusUnauthorized, w.Code)
}