package models

type UserResponse struct {
    ID    int    `json:"id"`
    Email string `json:"email"`
    Time  string `json:"time,omitempty"`
}

type UserRequest struct {
    ID       string `json:"id,omitempty"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Time     string `json:"time,omitempty"`
}

type LoginResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token,omitempty"`
}