package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Woun1zoN/go-identity-service/internal/error_handling"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
    UserIDKey contextKey = "user_id"
    RoleKey   contextKey = "role"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
        if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
            errorhandling.Unauthorized(w, r, GetRequestID(r))
            return
        }

        tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

        token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method")
            }
            return jwtKey, nil
        })

        if err != nil || !token.Valid {
            errorhandling.Unauthorized(w, r, GetRequestID(r))
            return
        }

        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok || claims["user_id"] == nil {
            errorhandling.Unauthorized(w, r, GetRequestID(r))
            return
        }

        userID := fmt.Sprintf("%v", claims["user_id"])

        ctx := context.WithValue(r.Context(), UserIDKey, userID)
        ctx = context.WithValue(ctx, RoleKey, claims["role"])

        next.ServeHTTP(w, r.WithContext(ctx))
	})
}