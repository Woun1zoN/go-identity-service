package migrations_test

import (
	"context"
	"testing"
	"fmt"
	"os"
	"strings"
	"path/filepath"
	

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

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
	
	fmt.Printf("DB_HOST from env: '%s'\n", os.Getenv("DB_HOST"))
	
	dbURL := migrations.BuildDBURL()
	fmt.Printf("Built URL: %s\n", dbURL)
	
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	
	migrationsPath := "file://" + strings.ReplaceAll(filepath.Join(wd, "..", "..", "internal/db/migrations"), "\\", "/")

	err = migrations.RunMigrations(dbURL, migrationsPath)
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	dbServer, err := db.InitDB(context.Background())
	if err != nil {
		t.Fatalf("failed to connect to DB: %v", err)
	}
	defer dbServer.DB.Close()

	err = CheckTables(dbServer.DB)
	if err != nil {
		t.Fatalf("tables check failed: %v", err)
	}
}