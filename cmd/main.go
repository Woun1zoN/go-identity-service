package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/Woun1zoN/go-identity-service/internal/app"
	"github.com/Woun1zoN/go-identity-service/internal/server"
	"github.com/Woun1zoN/go-identity-service/internal/db/migrations"
)

func main() {
	key := os.Getenv("JWT_SECRET")
	if key == "" {
		log.Fatal("JWT_SECRET not set")
	}

	dbURL := migrations.BuildDBURL(false)
	router, dbServer, err := app.InitApp(dbURL, []byte(key), "redis:6379")
	if err != nil {
		log.Fatal(err)
	}
	defer dbServer.DB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	server.Run(ctx, router)
}