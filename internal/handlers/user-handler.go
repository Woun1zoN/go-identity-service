package handlers

import (
	"net/http"
	"encoding/json"
	"strconv"

	"github.com/Woun1zoN/go-identity-service/internal/error_handling"
	"github.com/Woun1zoN/go-identity-service/internal/middleware"
)

func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
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

	response, err := h.Service.GetUserByID(r.Context(), userID)
	if errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
        return
	}

	json.NewEncoder(w).Encode(response)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
        return
	}

	response, err := h.Service.Refresh(r.Context(), req.RefreshToken, w, r)
	if errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
		return
	}

	json.NewEncoder(w).Encode(response)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
		return
	}

	if err := h.Service.Logout(r.Context(), req.RefreshToken); errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Successfully logged out",
	})
}

func (h *Handler) PromoteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	var req struct {UserID int `json:"user_id"`}

	if err := json.NewDecoder(r.Body).Decode(&req); errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
		return
	}

	if err := h.Service.PromoteUser(r.Context(), req.UserID); errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User promoted",
	})
}