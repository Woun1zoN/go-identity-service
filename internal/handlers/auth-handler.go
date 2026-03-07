package handlers

import (
	"net/http"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/Woun1zoN/go-identity-service/internal/models"
	"github.com/Woun1zoN/go-identity-service/internal/error_handling"
	"github.com/Woun1zoN/go-identity-service/internal/middleware"
	"github.com/Woun1zoN/go-identity-service/internal/auth"

	"golang.org/x/crypto/bcrypt"

	"github.com/jackc/pgx/v5"
)

func (Server *Handler) Registration(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	var input models.UserRequest
	if err := json.NewDecoder(r.Body).Decode(&input); errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
		return
	}

	if err := Server.Validate.Struct(input); errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
		return
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
		return
	}

	userID, err := Server.DB.CreateUser(r.Context(), input.Email, string(passHash))
	if errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
		return
	}

	response := models.UserResponse{
		ID:    userID,
		Email: input.Email,
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}

func (Server *Handler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	var input models.UserRequest
	if err := json.NewDecoder(r.Body).Decode(&input); errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
		return
	}

	if err := Server.Validate.Struct(input); errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
		return
	}

	user, err := Server.DB.GetUserByEmail(r.Context(), input.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		errorhandling.Unauthorized(w, r, middleware.GetRequestID(r))
		return
	} else if errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		errorhandling.Unauthorized(w, r, middleware.GetRequestID(r))
		return
	}

	accessToken, err := auth.GenerateAccessToken(strconv.Itoa(user.ID))
	if errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
		return
	}

	refreshToken, refreshID, refreshHash, err := auth.GenerateRefreshToken(strconv.Itoa(user.ID))
	if errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
		return
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err := Server.DB.InsertRefreshToken(r.Context(), refreshID, strconv.Itoa(user.ID), refreshHash, expiresAt); errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
		return
	}

	response := models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	_ = json.NewEncoder(w).Encode(response)
}