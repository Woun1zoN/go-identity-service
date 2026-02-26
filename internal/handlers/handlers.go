package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/Woun1zoN/go-identity-service/internal/auth"
	"github.com/Woun1zoN/go-identity-service/internal/db"
	"github.com/Woun1zoN/go-identity-service/internal/models"

	"golang.org/x/crypto/bcrypt"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
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
	defer r.Body.Close()

	var input models.UserRequest

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = Server.Validate.Struct(input)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	pass, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

    var userID int

	err = Server.DB.QueryRow(r.Context(), "INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id", input.Email, string(pass)).Scan(&userID)
	if err != nil {
        log.Println("DB error:", err)
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
	}

	response := models.RegisterResponse{
		ID:    userID,
		Email: input.Email,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (Server *DBHandler) Login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var input models.UserRequest

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = Server.Validate.Struct(input)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var pass string
	var userID int

	err = Server.DB.QueryRow(r.Context(), "SELECT id, password_hash FROM users WHERE email = $1", input.Email).Scan(&userID, &pass)
	if err != nil {
		if err == pgx.ErrNoRows {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
			log.Println(err)
            return
        }
        log.Println("DB error:", err)
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
	}

	err = bcrypt.CompareHashAndPassword([]byte(pass), []byte(input.Password))
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	} else {
		w.WriteHeader(http.StatusOK)
		token, err := auth.GenerateToken(strconv.Itoa(userID))
        if err != nil {
            panic(err)
        }

        log.Println("JWT:", token)
	}
}