package middleware

import (
	"context"
	"net"
	"net/http"
	"time"
    "fmt"
    "strings"

	"github.com/Woun1zoN/go-identity-service/internal/error_handling"

	"github.com/redis/go-redis/v9"
)

func AllowRequest(ctx context.Context, rdb *redis.Client, userID string, limit int, window time.Duration) (bool, error) {
    now := time.Now().Unix()
    key := "request:" + userID

    if err := rdb.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", now-int64(window.Seconds()))).Err(); err != nil {
        return false, err
    }

    count, err := rdb.ZCard(ctx, key).Result()
    if err != nil {
        return false, err
    }

    if int(count) >= limit {
        return false, nil
    }

    added, err := rdb.ZAddNX(ctx, key, redis.Z{Score: float64(now), Member: now}).Result()
    if err != nil {
        return false, err
    }

    if added > 0 {
        rdb.Expire(ctx, key, window)
    }

    return true, nil
}

func RateLimit(next http.Handler, limit int, window time.Duration, rdb *redis.Client) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip := strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]
        if ip == "" {
            ip, _, _ = net.SplitHostPort(r.RemoteAddr)
        }
        ip = strings.TrimSpace(ip)
    
        userID := ip
        if uid := r.Context().Value(UserIDKey); uid != nil {
            userID = fmt.Sprintf("%v", uid)
        }
        allowed, err := AllowRequest(r.Context(), rdb, userID, limit, window)

        if errorhandling.HTTPErrors(w, err, GetRequestID(r)) {
            return
        }

        if !allowed {
            err := errorhandling.ErrTooManyRequests
            errorhandling.HTTPErrors(w, err, GetRequestID(r))
            return
        }

        next.ServeHTTP(w, r)
    })
}

func RateLimiter(limit int, window time.Duration, rdb *redis.Client) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return RateLimit(next, limit, window, rdb)
    }
}