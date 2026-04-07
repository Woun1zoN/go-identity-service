package migrations_test

import (
	"context"
	"testing"
	"fmt"
	"os"
	"path/filepath"
	

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/Woun1zoN/go-identity-service/internal/db"
	"github.com/Woun1zoN/go-identity-service/internal/db/migrations"
)

func CheckTables(db *pgxpool.Pool) error {
	ctx := context.Background()
	var exists bool

	err := db.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name=$1)", "users").Scan(&exists)
    if err != nil {
        return fmt.Errorf("query failed for table users: %w", err)
    }
    if !exists {
        return fmt.Errorf("table users does not exist")
    }

	err = db.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name=$1)", "refresh_tokens").Scan(&exists)
	if err != nil {
        return fmt.Errorf("query failed for table refresh_tokens: %w", err)
    }
    if !exists {
        return fmt.Errorf("table refresh_tokens does not exist")
    }

	return nil
}

func TestMigrations(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to load .env: %v", err)
	}

	dbName := os.Getenv("DB_NAME_TEST")

	if err := migrations.SetupTestDB(dbName); err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}

	dbURL := migrations.BuildDBURL(true)

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	migrationsPath := "file://" + filepath.ToSlash(filepath.Join(wd, "..", "..", "internal/db/migrations"))

	if err := migrations.RunMigrations(dbURL, migrationsPath); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	dbServer, err := db.InitDB(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("failed to connect to DB: %v", err)
	}
	defer dbServer.DB.Close()

	if err := CheckTables(dbServer.DB); err != nil {
		t.Fatalf("tables check failed: %v", err)
	}

	m, err := migrate.New(migrationsPath, dbURL)
	if err != nil {
		t.Fatalf("failed to init migrate: %v", err)
	}
	defer m.Close()

	if err := m.Drop(); err != nil {
		t.Fatalf("failed to drop db: %v", err)
	}
}