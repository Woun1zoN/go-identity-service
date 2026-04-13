package integration_test

import (
	"testing"
	"net/http/httptest"
	"strings"
	"net/http"
	"encoding/json"
	"context"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"

	"github.com/Woun1zoN/go-identity-service/internal/app"
	"github.com/Woun1zoN/go-identity-service/internal/models"
	"github.com/Woun1zoN/go-identity-service/internal/db/migrations"
)

func TestRegisterUser(t *testing.T) {
	if err := godotenv.Load("../../.env"); err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to load .env: %v", err)
	}

	projectRoot := "../../"
	migrationsPath := "file://" + filepath.ToSlash(filepath.Join(projectRoot, "internal/db/migrations"))
	dbURL := migrations.BuildDBURL(true)

	if err := migrations.RunMigrations(dbURL, migrationsPath); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	router, dbServer, err := app.InitApp(dbURL, []byte("testkey"), "localhost:6379", true)
    require.NoError(t, err)
    defer dbServer.DB.Close()

	body := `{"email":"test@example.com","password":"supersecret"}`
	r := httptest.NewRequest("POST", "/register", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	require.Equal(t, http.StatusCreated, w.Code, "body: %s", w.Body.String())

	var response models.UserResponse
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	require.Equal(t, "test@example.com", response.Email)
	require.NotZero(t, response.ID)

	var email string
	err = dbServer.DB.QueryRow(context.Background(), "SELECT email FROM users WHERE id=$1", response.ID).Scan(&email)

    require.NoError(t, err)
	require.Equal(t, "test@example.com", email)
}