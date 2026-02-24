package handlers

import (
	"encoding/json"
	"net/http"
	"log"

	"github.com/Woun1zoN/go-identity-service/internal/db"
	"github.com/Woun1zoN/go-identity-service/internal/models"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBHandler struct {
	DB *pgxpool.Pool
	Validate *validator.Validate
}

func NewDBHandler(dbServer *db.DBServer, validate *validator.Validate) *DBHandler {
	return &DBHandler{
		DB: dbServer.DB,
		Validate: validate,
	}
}

func (Server *DBHandler) Registration(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

	var input models.RegisterRequest

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err = Server.Validate.Struct(input)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var userID int

	err = Server.DB.QueryRow(r.Context(), "INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id", input.Email, input.Password).Scan(&userID)
	if err != nil {
        log.Println("DB error:", err)
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}