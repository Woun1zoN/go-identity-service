package integration_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/Woun1zoN/go-identity-service/internal/auth"
	"github.com/Woun1zoN/go-identity-service/internal/db/migrations"
	"github.com/Woun1zoN/go-identity-service/internal/handlers"
	"github.com/Woun1zoN/go-identity-service/internal/middleware"
	"github.com/Woun1zoN/go-identity-service/internal/repository"
	"github.com/Woun1zoN/go-identity-service/internal/service"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var (
	testDB     *pgxpool.Pool
	TestDBURL  string
	TestRouter *chi.Mux
)

func TestMain(m *testing.M) {
	_ = godotenv.Load("../../.env.test")

	TestDBURL = migrations.BuildDBURL(true)

	if err := migrations.SetupTestDB(os.Getenv("DB_NAME_TEST")); err != nil {
		log.Fatal(err)
	}

	if err := migrations.RunMigrations(TestDBURL, "file://../../internal/db/migrations"); err != nil {
		log.Fatal(err)
	}

	pool, err := pgxpool.New(context.Background(), TestDBURL)
	if err != nil {
		log.Fatal(err)
	}

	testDB = pool

	TestRouter = setupRouter(testDB)

	code := m.Run()
	testDB.Close()
	os.Exit(code)
}

func setupRouter(dbPool *pgxpool.Pool) *chi.Mux {
	r := chi.NewRouter()
	validate := validator.New()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recovery)
	r.Use(middleware.Logger)
	r.Use(middleware.Context)

	jwtKey := []byte("testkey")
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	authService := &auth.AuthConfig{JWTKey: jwtKey}

	userRepo := repository.NewUserRepository(dbPool)
	svc := service.NewService(userRepo, authService)
	handler := handlers.NewHandler(svc, validate)

	r.Post("/logout", handler.Logout)
	r.With(middleware.RateLimiter(3, 0, rdb)).Post("/register", handler.Registration)
	r.With(middleware.RateLimiter(5, 0, rdb)).Post("/login", handler.Login)
	r.With(middleware.RateLimiter(10, 0, rdb)).Post("/refresh", handler.Refresh)
	r.With(middleware.RateLimiter(1, 0, rdb)).With(middleware.Auth(authService)).With(middleware.RequireRole("admin")).Post("/admin/promote", handler.PromoteUser)
	r.With(middleware.Auth(authService)).Get("/profile", handler.Profile)

	return r
}

func CleanDB(t *testing.T) {
	_, err := testDB.Exec(context.Background(),
		`TRUNCATE TABLE refresh_tokens CASCADE;
		TRUNCATE TABLE users CASCADE;`,
	)
	if err != nil {
		t.Fatalf("Failed to clean DB: %v", err)
	}
}