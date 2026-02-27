package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Woun1zoN/go-identity-service/internal/db"
	"github.com/Woun1zoN/go-identity-service/internal/handlers"
	"github.com/Woun1zoN/go-identity-service/internal/middleware"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

func main() {
	// Variables

	r := chi.NewRouter()
	ctx := context.Background()
	validate := validator.New()
	godotenv.Load()

	// Middleware

	// Connection DB

	dbServer, err := db.InitDB(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer dbServer.DB.Close()

	log.Println("Подключен к БД")

	dbHandler := handlers.NewDBHandler(dbServer, validate)

	// Handlers

	r.Post("/register", dbHandler.Registration)
	r.Post("/login", dbHandler.Login)

	r.With(middleware.Auth).Get("/profile", dbHandler.Profile)

	// Starting

	log.Println("Сервер запущен на http://localhost:8080")
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal("Сервер словил грустного:", err)
	}
}