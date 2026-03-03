package middleware

import (
	"context"
	"time"
	"net/http"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

var rdb = redis.NewClient(&redis.Options{
    Addr:     "localhost:6380",
    Password: "",
    DB:       0,
})

func AllowRequest(userID string, limit int, window time.Duration) (bool, error) {
    key := "rate:" + userID
    val, err := rdb.Incr(ctx, key).Result()
    if err != nil {
        return false, err
    }
    if val == 1 {
        rdb.Expire(ctx, key, window)
    }
    if val > int64(limit) {
        return false, nil
    }
    return true, nil
}

func RateLimit(next http.Handler, limit int, window time.Duration) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := r.RemoteAddr
        allowed, err := AllowRequest(userID, limit, window)
        if err != nil {
            http.Error(w, "Internal Server Error", 500)
            return
        }
        if !allowed {
            http.Error(w, "Too Many Requests", 429)
            return
        }
        next.ServeHTTP(w, r)
    })
}