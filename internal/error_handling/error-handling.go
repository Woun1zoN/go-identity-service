package errorhandling

import (
	"net/http"
	"errors"
	"encoding/json"
	"context"
	"io"
	"strconv"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type MapHTTPError struct {
	Error   string
	Status 	int
	Message string
}

type ErrorResponse struct {
    Error     string `json:"error"`
    Message   string `json:"message"`
    RequestID string `json:"request_id"`
}

var Errors = map[string]MapHTTPError {
	"Timeout":         {Error: "request_timeout",       Status: http.StatusRequestTimeout,      Message: http.StatusText(http.StatusRequestTimeout)},
	"Internal":        {Error: "internal_server_error", Status: http.StatusInternalServerError, Message: http.StatusText(http.StatusInternalServerError)},
	"BadJSON":         {Error: "bad_json",              Status: http.StatusBadRequest,          Message: http.StatusText(http.StatusBadRequest)},
	"ErrorValidation": {Error: "error_validation",      Status: http.StatusBadRequest,          Message: http.StatusText(http.StatusBadRequest)},
	"NotFound":        {Error: "not_found",             Status: http.StatusNotFound,            Message: http.StatusText(http.StatusNotFound)},
	"Unauthorized":    {Error: "unauthorized",          Status: http.StatusUnauthorized,        Message: http.StatusText(http.StatusUnauthorized)},
	"TooManyRequests": {Error: "too_many_requests",     Status: http.StatusTooManyRequests,     Message: http.StatusText(http.StatusTooManyRequests)},
	"Conflict":        {Error: "conflict",              Status: http.StatusConflict,            Message: http.StatusText(http.StatusConflict)},
    "Forbidden":       {Error: "forbidden",             Status: http.StatusForbidden,           Message: http.StatusText(http.StatusForbidden)},
}

func ErrorEncoding(w http.ResponseWriter, status int, err string, message string, req_id string) {
	w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)

	errResp := ErrorResponse{
		Error: err,
		Message: message,
		RequestID: req_id,
	}

	json.NewEncoder(w).Encode(errResp)
}

func Unauthorized(w http.ResponseWriter, r *http.Request, id string) {
    ErrorEncoding(
		w,
		Errors["Unauthorized"].Status,
		Errors["Unauthorized"].Error,
		"Unauthorized access",
		id,
	)
}

func Forbidden(w http.ResponseWriter, r *http.Request, id string) {
    ErrorEncoding(
        w,
        Errors["Forbidden"].Status,
        Errors["Forbidden"].Error,
        "Insufficient permissions",
        id,
    )
}

func HTTPErrors(w http.ResponseWriter, err error, id string) bool {
    if err == nil {
        return false
    }

    var syntaxErr *json.SyntaxError
    var typeErr *json.UnmarshalTypeError

    if errors.As(err, &syntaxErr) {
        log.Printf("Error [%s] | %v", id, err)
        ErrorEncoding(w, Errors["BadJSON"].Status, Errors["BadJSON"].Error, "Invalid JSON syntax", id)
        return true
    }
    if errors.As(err, &typeErr) {
        log.Printf("Error [%s] | %v", id, err)
        ErrorEncoding(w, Errors["BadJSON"].Status, Errors["BadJSON"].Error, "JSON type mismatch", id)
        return true
    }

    if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
        log.Printf("Error [%s] | %v", id, err)
        ErrorEncoding(w, Errors["Conflict"].Status, Errors["Conflict"].Error, "Data conflict", id)
        return true
    }

    switch {
    case errors.Is(err, context.DeadlineExceeded):
        log.Printf("Error [%s] | The request timeout has expired: %v", id, err)
        ErrorEncoding(w, Errors["Timeout"].Status, Errors["Timeout"].Error, "The request timeout has expired", id)
        return true

    case errors.Is(err, pgx.ErrNoRows),
		errors.Is(err, ErrNoRowsAffected):
        log.Printf("Error [%s] | Not found: %v", id, err)
        ErrorEncoding(w, Errors["NotFound"].Status, Errors["NotFound"].Error, "Not found", id)
        return true

    case errors.Is(err, io.EOF):
        log.Printf("Error [%s] | Empty request body: %v", id, err)
        ErrorEncoding(w, Errors["BadJSON"].Status, Errors["BadJSON"].Error, "Empty request body", id)
        return true

    case errors.Is(err, strconv.ErrSyntax), errors.Is(err, strconv.ErrRange):
        log.Printf("Error [%s] | Invalid input: %v", id, err)
        ErrorEncoding(w, Errors["ErrorValidation"].Status, Errors["ErrorValidation"].Error, "Invalid input", id)
        return true

    case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword),
        errors.Is(err, jwt.ErrTokenMalformed),
        errors.Is(err, jwt.ErrTokenExpired),
        errors.Is(err, jwt.ErrTokenNotValidYet):
        log.Printf("Error [%s] | Unauthorized access: %v", id, err)
        ErrorEncoding(w, Errors["Unauthorized"].Status, Errors["Unauthorized"].Error, "Unauthorized access", id)
        return true

    case errors.Is(err, ErrTooManyRequests):
        log.Printf("Error [%s] | The request limit has been exceeded: %v", id, err)
        ErrorEncoding(w, Errors["TooManyRequests"].Status, Errors["TooManyRequests"].Error, "The request limit has been exceeded", id)
        return true

    default:
        log.Printf("Error [%s] | Internal server error: %v", id, err)
        ErrorEncoding(w, Errors["Internal"].Status, Errors["Internal"].Error, "Internal server error", id)
        return true
    }
}