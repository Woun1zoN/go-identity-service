package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DBServer struct {
	DB *pgxpool.Pool
}

func InitDB(ctx context.Context, dbURL string) (*DBServer, error) {
	pool, err := pgxpool.New(ctx, dbURL)

	if err != nil {
		return nil, fmt.Errorf("No connection to DB: %w", err)
	}

	ServerDB := &DBServer{
		DB: pool,
	}

	return ServerDB, nil
}