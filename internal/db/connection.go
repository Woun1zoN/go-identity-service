package db

import (
	"context"
	"os"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DBServer struct {
	DB *pgxpool.Pool
}

func InitDB(ctx context.Context) (*DBServer, error) {
	user := os.Getenv("DB_USER")
    password := os.Getenv("DB_PASSWORD")
    dbname := os.Getenv("DB_NAME")

	conn := fmt.Sprintf(
    "postgres://%s:%s@db:5432/%s?sslmode=disable",
    user, password, dbname,
)

	pool, err := pgxpool.New(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("No connection to DB: %w", err)
	}

	ServerDB := &DBServer{
		DB: pool,
	}

	return ServerDB, nil
}