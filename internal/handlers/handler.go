package handlers

import (
	"github.com/Woun1zoN/go-identity-service/internal/repository"

	"github.com/go-playground/validator/v10"
)

type Handler struct {
	DB *repository.UserRepository
	Validate *validator.Validate
}

func NewHandler(userRepo *repository.UserRepository, validate *validator.Validate) *Handler {
	return &Handler{
		DB: userRepo,
		Validate: validate,
	}
}