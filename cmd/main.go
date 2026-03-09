package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Woun1zoN/go-identity-service/internal/db"
	"github.com/Woun1zoN/go-identity-service/internal/repository"
	"github.com/Woun1zoN/go-identity-service/internal/handlers"
	"github.com/Woun1zoN/go-identity-service/internal/middleware"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Variables

	r := chi.NewRouter()
	ctx := context.Background()
	validate := validator.New()
	godotenv.Load()

	rdb := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
    })

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

	log.Println("Подключен к БД")

	userRepo := repository.NewUserRepository(dbServer.DB)
    handler := handlers.NewHandler(userRepo, validate)

	// Handlers

	r.Post("/register", handler.Registration)
	r.Post("/logout", handler.Logout)

	limit := 5
    window := time.Minute

	r.With(func(next http.Handler) http.Handler {
        return middleware.RateLimit(next, limit, window, rdb)
	}).Post("/login", handler.Login)

	r.With(func(next http.Handler) http.Handler {
        return middleware.RateLimit(next, limit, window, rdb)
    }).Post("/refresh", handler.Refresh)

	r.With(middleware.Auth).Get("/profile", handler.Profile)

	// Starting

	log.Println("Сервер запущен на http://localhost:8080")
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal("Сервер словил грустного:", err)
	}
}