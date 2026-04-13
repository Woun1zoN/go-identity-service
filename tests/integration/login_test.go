package integration_test

import (
	"context"
	"testing"
	"os"
	"path/filepath"
	"net/http/httptest"
	"strings"
	"net/http"
	"encoding/json"
	"strconv"
	"fmt"

	"github.com/stretchr/testify/require"
	"github.com/joho/godotenv"
	"github.com/golang-jwt/jwt/v5"

	"github.com/Woun1zoN/go-identity-service/internal/auth"
	"github.com/Woun1zoN/go-identity-service/internal/models"
	"github.com/Woun1zoN/go-identity-service/internal/repository"
	"github.com/Woun1zoN/go-identity-service/internal/db/migrations"
	"github.com/Woun1zoN/go-identity-service/internal/app"
)

var testKey = []byte("testkey")

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
	if err := godotenv.Load("../../.env"); err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to load .env: %v", err)
	}

	projectRoot := "../../"
	migrationsPath := "file://" + filepath.ToSlash(filepath.Join(projectRoot, "internal/db/migrations"))
	dbURL := migrations.BuildDBURL(true)

	if err := migrations.RunMigrations(dbURL, migrationsPath); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	router, dbServer, err := app.InitApp(dbURL, testKey, "localhost:6379", true)
    if err != nil {
        t.Fatal(err)
    }
    defer dbServer.DB.Close()

	CleanDB(t, dbServer.DB)

	email := "test@example.com"
	password := "supersecret"

	hash, _ := auth.HashPassword(password)

	user := models.User{
		Email:        email,
		PasswordHash: hash,
	}

	repo := repository.NewUserRepository(dbServer.DB)

	userID, err := repo.CreateUser(context.Background(), user.Email, user.PasswordHash)
	require.NoError(t, err)

	body := fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, password)
	r := httptest.NewRequest("POST", "/login", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

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