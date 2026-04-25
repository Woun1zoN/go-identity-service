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
    var user, password, dbname string

    if test {
        user = os.Getenv("DB_USER_TEST")
        password = os.Getenv("DB_PASSWORD_TEST")
        dbname = os.Getenv("DB_NAME_TEST")
    } else {
        user = os.Getenv("DB_USER")
        password = os.Getenv("DB_PASSWORD")
        dbname = os.Getenv("DB_NAME")
    }

    host := os.Getenv("DB_HOST")
    if host == "" {
        host = "localhost"
    }

    port := os.Getenv("DB_PORT")
    if port == "" {
        if test {
            port = "5433"
        } else {
            port = "5432"
        }
    }

    return fmt.Sprintf(
        "postgres://%s:%s@%s:%s/%s?sslmode=disable",
        user, password, host, port, dbname,
    )
}

func SetupTestDB(dbName string) error {
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER_TEST")
	password := os.Getenv("DB_PASSWORD_TEST")
	port := os.Getenv("DB_PORT")

	if user == "" || password == "" {
        return fmt.Errorf("test env not loaded properly")
	}

	if host == "" {
		host = "localhost"
	}

	if port == "" {
    if os.Getenv("TEST_ENV") == "1" {
        port = "5433"
    } else {
        port = "5432"
    }
	}

	url := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=disable", user, password, host, port)
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