package main

import (
	"context"
	"log"
	"os"
	"time"
	"strings"
    "path/filepath"

	"github.com/Woun1zoN/go-identity-service/internal/auth"
	"github.com/Woun1zoN/go-identity-service/internal/db"
	"github.com/Woun1zoN/go-identity-service/internal/db/migrations"
	"github.com/Woun1zoN/go-identity-service/internal/handlers"
	"github.com/Woun1zoN/go-identity-service/internal/middleware"
	"github.com/Woun1zoN/go-identity-service/internal/repository"
	"github.com/Woun1zoN/go-identity-service/internal/server"
	"github.com/Woun1zoN/go-identity-service/internal/service"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Variables

	r := chi.NewRouter()
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	validate := validator.New()

	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatal("Have little problems...")
	}

	key := os.Getenv("JWT_SECRET")
	if key == "" {
		log.Fatal("JWT_SECRET not set")
	}

	jwtKey := []byte(key)
	authService := &auth.AuthConfig{JWTKey: jwtKey}

	rdb := redis.NewClient(&redis.Options{
        Addr: "redis:6379",
        Password: "",
        DB:       0,
    })
	defer rdb.Close()

	// Middleware

	r.Use(middleware.RequestID)
	r.Use(middleware.Recovery)
	r.Use(middleware.Logger)
	r.Use(middleware.Context)

	// Connection DB

	dbURL := migrations.BuildDBURL(false)

	dbServer, err := db.InitDB(context.Background(), dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer dbServer.DB.Close()

	log.Println("Connected to DB")

	wd, err := os.Getwd()
	if err != nil {
        log.Fatal(err)
	}
	migrationsPath := "file://" + strings.ReplaceAll(filepath.Join(wd, "internal/db/migrations"), "\\", "/")
	err = migrations.RunMigrations(migrations.BuildDBURL(false), migrationsPath)
    if err != nil {
		log.Fatalf("Migration error: %v", err)
    }

	userRepo := repository.NewUserRepository(dbServer.DB)
	service := service.NewService(userRepo, authService)
    handler := handlers.NewHandler(service, validate)

	// Handlers

	r.Post("/logout", handler.Logout)

	r.With(middleware.RateLimiter(3, time.Minute, rdb)).Post("/register", handler.Registration)
	r.With(middleware.RateLimiter(5, time.Minute, rdb)).Post("/login", handler.Login)
	r.With(middleware.RateLimiter(10, time.Minute, rdb)).Post("/refresh", handler.Refresh)

	r.With(middleware.RateLimiter(1, time.Minute, rdb)).With(middleware.Auth(authService)).With(middleware.RequireRole("admin")).Post("/admin/promote", handler.PromoteUser)
	r.With(middleware.Auth(authService)).Get("/profile", handler.Profile)

	// Starting

	server.Run(ctx, r)
}