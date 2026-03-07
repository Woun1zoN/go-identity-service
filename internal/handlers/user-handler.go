package handlers

import (
	"net/http"
	"encoding/json"
	"errors"
	"strconv"
	"time"
	"fmt"

	"github.com/Woun1zoN/go-identity-service/internal/models"
	"github.com/Woun1zoN/go-identity-service/internal/error_handling"
	"github.com/Woun1zoN/go-identity-service/internal/middleware"
	"github.com/Woun1zoN/go-identity-service/internal/auth"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
)

func (Server *Handler) Profile(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    userIDStr, ok := r.Context().Value(middleware.UserIDKey).(string)
    if !ok {
        errorhandling.Unauthorized(w, r, middleware.GetRequestID(r))
        return
    }
	userID, err := strconv.Atoi(userIDStr)
    if errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
        return
	}

	response, err := Server.DB.GetUserByID(r.Context(), userID)
	if errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
        return
	}

	json.NewEncoder(w).Encode(response)
}

func (Server *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
        return
	}

	token, err := jwt.Parse(req.RefreshToken, func(t *jwt.Token) (interface{}, error) {
    if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("unexpected signing method")
    }
    return auth.JwtKey, nil
    })

    if errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
		return
    }

	if !token.Valid {
        errorhandling.Unauthorized(w, r, middleware.GetRequestID(r))
        return
    }

	claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        errorhandling.Unauthorized(w, r, middleware.GetRequestID(r))
        return
    }
    jti, ok := claims["jti"].(string)
    if !ok {
        errorhandling.Unauthorized(w, r, middleware.GetRequestID(r))
        return
    }
    userID, ok := claims["user_id"].(string)
    if !ok {
        errorhandling.Unauthorized(w, r, middleware.GetRequestID(r))
        return
    }

	hash := auth.HashToken(req.RefreshToken)

	storedToken, err := Server.DB.GetRefreshTokenByID(r.Context(), jti)
    if errors.Is(err, pgx.ErrNoRows) {
		errorhandling.Unauthorized(w, r, middleware.GetRequestID(r))
		return
    }

    if errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
        return
    }

	if storedToken.TokenHash != hash || storedToken.Revoked || time.Now().After(storedToken.ExpiresAt) {
        errorhandling.Unauthorized(w, r, middleware.GetRequestID(r))
        return
    }

	if err := Server.DB.RevokeRefreshToken(r.Context(), jti); errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
		return
	}

	refresh, newJTI, _, err := auth.GenerateRefreshToken(userID)
	if errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
		return
	}

	newHash := auth.HashToken(refresh)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	if err := Server.DB.InsertRefreshToken(r.Context(), newJTI, userID, newHash, expiresAt); errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
		return
	}

	access, err := auth.GenerateAccessToken(userID)
	if errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
		return
	}

	response := models.TokenResponse{
		AccessToken: access,
		RefreshToken: refresh,
	}

	json.NewEncoder(w).Encode(response)
}