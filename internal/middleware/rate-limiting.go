package middleware

import (
	"context"
	"time"
	"net/http"
    "net"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func AllowRequest(userID string, limit int, window time.Duration, rdb *redis.Client) (bool, error) {
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

func RateLimit(next http.Handler, limit int, window time.Duration, rdb *redis.Client) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip, _, _ := net.SplitHostPort(r.RemoteAddr)
        userID := ip
        allowed, err := AllowRequest(userID, limit, window, rdb)
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