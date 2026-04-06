package migrations

import (
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(dbURL, migrationsPath string) error {
    m, err := migrate.New(migrationsPath, dbURL)
    if err != nil {
        return err
    }

    if err := m.Down(); err != nil && err != migrate.ErrNoChange {
        return err
    }

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return err
    }

    return nil
}

func BuildDBURL() string {
    user := os.Getenv("DB_USER")
    password := os.Getenv("DB_PASSWORD")
    dbname := os.Getenv("DB_NAME")
    host := os.Getenv("DB_HOST")

    if host == "" {
        host = "localhost"
    }

    return fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable",
        user, password, host, dbname,
    )
}