package migrations

import (
	"fmt"
	"os"
    "context"
    "log"
	"time"

	"github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
    "github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(dbURL, migrationsPath string) error {
	if err := WaitForDB(dbURL, 30*time.Second); err != nil {
		return fmt.Errorf("database not ready: %w", err)
	}

    m, err := migrate.New(migrationsPath, dbURL)
    if err != nil {
        return err
    }

    defer m.Close()

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return err
    }

	if err := WaitForTables(dbURL, 30*time.Second); err != nil {
		return fmt.Errorf("tables not ready: %w", err)
	}

    return nil
}

func BuildDBURL(test bool) string {
    user := os.Getenv("DB_USER")
    password := os.Getenv("DB_PASSWORD")
    host := os.Getenv("DB_HOST")

    if host == "" {
        host = "localhost"
    }

    dbname := os.Getenv("DB_NAME")
	if test {
		dbname = os.Getenv("DB_NAME_TEST")
	}

    return fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable",
        user, password, host, dbname,
    )
}

func SetupTestDB(dbName string) error {
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	if host == "" {
		host = "localhost"
	}

	url := fmt.Sprintf("postgres://%s:%s@%s:5432/postgres?sslmode=disable", user, password, host)
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer pool.Close()

	var exists bool
	err = pool.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)", dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check db existence: %w", err)
	}

	if !exists {
		_, err := pool.Exec(context.Background(), fmt.Sprintf("CREATE DATABASE %s", dbName))
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
		log.Println("Database created:", dbName)
	}

	return nil
}

func WaitForDB(dbURL string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return fmt.Errorf("create pool: %w", err)
	}
	defer pool.Close()

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	var lastErr error

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for db: %w", lastErr)
		case <-ticker.C:
			pingCtx, cancelPing := context.WithTimeout(context.Background(), time.Second)
			lastErr = pool.Ping(pingCtx)
			cancelPing()

			if lastErr == nil {
				return nil
			}
		}
	}
}

func WaitForTables(dbURL string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return fmt.Errorf("create pool: %w", err)
	}
	defer pool.Close()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	var lastErr error

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for tables: %w", lastErr)
		case <-ticker.C:
			pingCtx, cancelPing := context.WithTimeout(context.Background(), time.Second)

			var exists bool
			lastErr = pool.QueryRow(pingCtx,
				"SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name='users')").Scan(&exists)

			cancelPing()

			if lastErr == nil && exists {
				return nil
			}
		}
	}
}