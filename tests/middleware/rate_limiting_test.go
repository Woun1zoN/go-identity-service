package middleware_test

import (
	"context"
	"time"
	"testing"

	"github.com/redis/go-redis/v9"

	"github.com/Woun1zoN/go-identity-service/internal/middleware"
)

func TestAllowRequest(t *testing.T) {
	ctx := context.Background()
	fakeNow := int64(1_000_000)

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB: 1,
	})

	if err := rdb.FlushDB(ctx).Err(); err != nil {
        t.Fatal("failed to flush redis:", err)
	}

	userID := "127.0.0.1"
	limit := 2
	window := time.Minute

	ok, err := middleware.AllowRequest(ctx, rdb, userID, limit, window, func() int64 { return fakeNow + 1 })
	if err != nil || !ok {
        t.Fatal("expected first request to pass")
	}

    ok, _ = middleware.AllowRequest(ctx, rdb, userID, limit, window, func() int64 { return fakeNow + 2 })
    if !ok {
        t.Fatal("expected second request to pass")
    }

    ok, _ = middleware.AllowRequest(ctx, rdb, userID, limit, window, func() int64 { return fakeNow + 3 })
    if ok {
        t.Fatalf("expected third request to be blocked")
    }

	ok, _ = middleware.AllowRequest(ctx, rdb, userID, limit, window, func() int64 { return fakeNow + int64(window.Seconds()) + 4 })
	if !ok {
		t.Fatal("expected request after window to pass")
	}
}