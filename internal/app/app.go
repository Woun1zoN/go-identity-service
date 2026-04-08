package app

import (
	"context"
	"strings"
	"path/filepath"
	"os"

	"github.com/Woun1zoN/go-identity-service/internal/auth"
	"github.com/Woun1zoN/go-identity-service/internal/db"
	"github.com/Woun1zoN/go-identity-service/internal/db/migrations"
	"github.com/Woun1zoN/go-identity-service/internal/handlers"
	"github.com/Woun1zoN/go-identity-service/internal/middleware"
	"github.com/Woun1zoN/go-identity-service/internal/repository"
	"github.com/Woun1zoN/go-identity-service/internal/service"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
)

func InitApp(dbURL string, jwtKey []byte, redisAddr string) (*chi.Mux, *db.DBServer, error) {
	r := chi.NewRouter()
	validate := validator.New()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recovery)
	r.Use(middleware.Logger)
	r.Use(middleware.Context)

	dbServer, err := db.InitDB(context.Background(), dbURL)
	if err != nil {
		return nil, nil, err
	}

	wd, _ := os.Getwd()
	migrationsPath := "file://" + strings.ReplaceAll(filepath.Join(wd, "internal/db/migrations"), "\\", "/")
	if err := migrations.RunMigrations(dbURL, migrationsPath); err != nil {
		return nil, nil, err
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	authService := &auth.AuthConfig{JWTKey: jwtKey}

	userRepo := repository.NewUserRepository(dbServer.DB)
	svc := service.NewService(userRepo, authService)
	handler := handlers.NewHandler(svc, validate)

	r.Post("/logout", handler.Logout)
	r.With(middleware.RateLimiter(3, 0, rdb)).Post("/register", handler.Registration)
	r.With(middleware.RateLimiter(5, 0, rdb)).Post("/login", handler.Login)
	r.With(middleware.RateLimiter(10, 0, rdb)).Post("/refresh", handler.Refresh)
	r.With(middleware.RateLimiter(1, 0, rdb)).With(middleware.Auth(authService)).With(middleware.RequireRole("admin")).Post("/admin/promote", handler.PromoteUser)
	r.With(middleware.Auth(authService)).Get("/profile", handler.Profile)

	return r, dbServer, nil
}