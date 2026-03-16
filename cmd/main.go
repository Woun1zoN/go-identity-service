package main

import (
	"context"
	"log"
	"time"
	"os"

	"github.com/Woun1zoN/go-identity-service/internal/db"
	"github.com/Woun1zoN/go-identity-service/internal/repository"
	"github.com/Woun1zoN/go-identity-service/internal/handlers"
	"github.com/Woun1zoN/go-identity-service/internal/middleware"
	"github.com/Woun1zoN/go-identity-service/internal/auth"
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

	if err := auth.GetJWTKey(); err != nil {
        log.Fatal(err)
    }

	rdb := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
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

	dbServer, err := db.InitDB(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer dbServer.DB.Close()

	log.Println("Connected to DB")

	userRepo := repository.NewUserRepository(dbServer.DB)
	service := service.NewService(userRepo)
    handler := handlers.NewHandler(service, validate)

	// Handlers

	r.Post("/logout", handler.Logout)

	r.With(middleware.RateLimiter(3, time.Minute, rdb)).Post("/register", handler.Registration)
	r.With(middleware.RateLimiter(5, time.Minute, rdb)).Post("/login", handler.Login)
	r.With(middleware.RateLimiter(10, time.Minute, rdb)).Post("/refresh", handler.Refresh)

	r.With(middleware.RateLimiter(1, time.Minute, rdb)).With(middleware.Auth).With(middleware.RequireRole("admin")).Post("/admin/promote", handler.PromoteUser)
	r.With(middleware.Auth).Get("/profile", handler.Profile)

	// Starting

	server.Run(ctx, r)
}