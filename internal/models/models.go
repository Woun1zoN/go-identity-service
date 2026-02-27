package models

type RegisterResponse struct {
    ID    int    `json:"id"`
    Email string `json:"email"`
}

type UserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}

type LoginResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token,omitempty"`
}