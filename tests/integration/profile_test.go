package integration_test

import (
	"testing"
	"context"
	"net/http/httptest"
	"net/http"
	"encoding/json"

	"github.com/stretchr/testify/require"

	"github.com/Woun1zoN/go-identity-service/internal/repository"
	"github.com/Woun1zoN/go-identity-service/internal/auth"
	"github.com/Woun1zoN/go-identity-service/internal/models"
)

func TestProfile(t *testing.T) {
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

	token, err := authService.GenerateAccessToken(&models.User{
		ID: userID,
		Email: email,
		Role: "user",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()

	TestRouter.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp models.UserResponse
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	require.Equal(t, userID, resp.ID)
	require.Equal(t, email, resp.Email)
	require.Equal(t, "user", resp.Role)
}