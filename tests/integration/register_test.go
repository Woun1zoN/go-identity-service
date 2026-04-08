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
    if err != nil {
        t.Fatal(err)
    }
    defer dbServer.DB.Close()

	body := `{"email":"test@example.com","password":"supersecret"}`
	r := httptest.NewRequest("POST", "/register", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d, body: %s", w.Code, w.Body.String())
	}

	var response models.UserResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
        t.Fatal(err)
    }
	if response.Email != "test@example.com" || response.ID == 0 {
        t.Fatalf("unexpected response: %+v", response)
    }

	var email string
	err = dbServer.DB.QueryRow(context.Background(), "SELECT email FROM users WHERE id=$1", response.ID).Scan(&email)
    if err != nil || email != "test@example.com" {
        t.Fatalf("user not found in DB or email mismatch: %v", err)
    }
}