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

var ctx = context.Background()

func AllowRequest(rdb *redis.Client, userID string, limit int, window time.Duration) (bool, error) {
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
        allowed, err := AllowRequest(rdb, userID, limit, window)

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