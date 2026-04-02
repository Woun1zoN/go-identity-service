package middleware_test

import (
	"testing"
	"net/http/httptest"
	"net/http"
	"context"

	"github.com/Woun1zoN/go-identity-service/internal/middleware"
)

func TestRequireRole(t *testing.T) {
    rec := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/", nil)

    called := false
    next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        called = true
    })

    middleware.RequireRole("admin")(next).ServeHTTP(rec, req)
    if rec.Code != http.StatusUnauthorized && rec.Code != http.StatusForbidden {
        t.Fatalf("expected forbidden/unauthorized, got %d", rec.Code)
    }
    if called {
        t.Fatal("next should not be called")
    }

    rec = httptest.NewRecorder()
    ctx := context.WithValue(req.Context(), middleware.RoleKey, "admin")
    middleware.RequireRole("admin")(next).ServeHTTP(rec, req.WithContext(ctx))
    if !called {
        t.Fatal("next should be called")
    }
}