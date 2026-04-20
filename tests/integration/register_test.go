package integration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Woun1zoN/go-identity-service/internal/models"
)

func TestRegisterUser(t *testing.T) {
	CleanDB(t)

	body := `{"email":"test@example.com","password":"supersecret"}`
	r := httptest.NewRequest("POST", "/register", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	TestRouter.ServeHTTP(w, r)

	require.Equal(t, http.StatusCreated, w.Code, "body: %s", w.Body.String())

	var response models.UserResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	require.Equal(t, "test@example.com", response.Email)
	require.NotZero(t, response.ID)

	var email string
	err = testDB.QueryRow(context.Background(), "SELECT email FROM users WHERE id=$1", response.ID).Scan(&email)
	require.NoError(t, err)
	require.Equal(t, "test@example.com", email)
}