package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
	"fmt"

	"github.com/Woun1zoN/go-identity-service/internal/auth"
	"github.com/Woun1zoN/go-identity-service/internal/db"
	"github.com/Woun1zoN/go-identity-service/internal/models"

	"golang.org/x/crypto/bcrypt"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/golang-jwt/jwt/v5"
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

	response := models.UserResponse{
		ID:    userID,
		Email: input.Email,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (Server *DBHandler) Login(w http.ResponseWriter, r *http.Request) {
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

	var pass string
	var userID int

	err = Server.DB.QueryRow(r.Context(), "SELECT id, password_hash FROM users WHERE email = $1", input.Email).Scan(&userID, &pass)
	if err != nil {
		if err == pgx.ErrNoRows {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
		accessToken, err := auth.GenerateAccessToken(strconv.Itoa(userID))
        if err != nil {
            http.Error(w, "Failed to generate token", http.StatusInternalServerError)
            return
        }

		refreshToken, refreshID, refreshHash, err := auth.GenerateRefreshToken(strconv.Itoa(userID))
        if err != nil {
            http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
            return
        }

		_, err = Server.DB.Exec(r.Context(), "INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at) VALUES ($1, $2, $3, $4)", refreshID, userID, refreshHash, time.Now().Add(7*24*time.Hour))
		if err != nil {
            http.Error(w, "Failed to save refresh token", http.StatusInternalServerError)
            return
        }

        response := models.LoginResponse{
			AccessToken: accessToken,
			RefreshToken: refreshToken,
		}

		json.NewEncoder(w).Encode(response)
	}
}

func (Server *DBHandler) Profile(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    userIDStr := r.Context().Value("user_id").(string)
	userID, err := strconv.Atoi(userIDStr)
    if err != nil {
        http.Error(w, "Invalid user ID", http.StatusUnauthorized)
        return
    }

	var createdAt time.Time
	response := models.UserResponse{}

	err = Server.DB.QueryRow(r.Context(), "SELECT id, email, created_at FROM users WHERE id = $1", userID).Scan(&response.ID, &response.Email, &createdAt)
	if err != nil {
        http.Error(w, "You do not have an account, please log in or register", http.StatusUnauthorized)
		log.Println(err)
        return
	}

	response.Time = createdAt.Format("2006-01-02 15:04:05")

	json.NewEncoder(w).Encode(response)
}

func (Server *DBHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	token, err := jwt.Parse(req.RefreshToken, func(t *jwt.Token) (interface{}, error) {
    if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("unexpected signing method")
    }
    return auth.JwtKey, nil
    })
	if err != nil || !token.Valid {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

	claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    jti, ok := claims["jti"].(string)
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    userID, ok := claims["user_id"].(string)
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

	hash := auth.HashToken(req.RefreshToken)

	var revoked bool
    var expiresAt time.Time
	var token_hash string

	err = Server.DB.QueryRow(r.Context(), "SELECT token_hash, revoked, expires_at FROM refresh_tokens WHERE id = $1", jti).Scan(&token_hash, &revoked, &expiresAt)
	if err != nil || token_hash != hash || revoked || time.Now().After(expiresAt) {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

	access, err := auth.GenerateAccessToken(userID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
        return
	}
	json.NewEncoder(w).Encode(map[string]string{
        "access_token": access,
    })
}