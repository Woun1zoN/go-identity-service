package handlers

import (
	"net/http"
	"encoding/json"

	"github.com/Woun1zoN/go-identity-service/internal/models"
	"github.com/Woun1zoN/go-identity-service/internal/error_handling"
	"github.com/Woun1zoN/go-identity-service/internal/middleware"
	"github.com/Woun1zoN/go-identity-service/internal/service"

	"github.com/go-playground/validator/v10"
)

type Handler struct {
	Service     *service.Service
	Validate    *validator.Validate
}

func NewHandler(service *service.Service, validate *validator.Validate) *Handler {
	return &Handler{
		Service:     service,
		Validate:    validate,
	}
}

func (h *Handler) Registration(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	var input models.UserRequest
	if err := json.NewDecoder(r.Body).Decode(&input); errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
		return
	}

	if err := h.Validate.Struct(input); errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
		return
	}

	response, err := h.Service.RegisterUser(r.Context(), input.Email, input.Password)
	if errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	var input models.UserRequest
	if err := json.NewDecoder(r.Body).Decode(&input); errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
		return
	}

	if err := h.Validate.Struct(input); errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
		return
	}

	response, err := h.Service.Login(r.Context(), input.Email, input.Password)
	if errorhandling.HTTPErrors(w, err, middleware.GetRequestID(r)) {
		return
	}

	_ = json.NewEncoder(w).Encode(response)
}