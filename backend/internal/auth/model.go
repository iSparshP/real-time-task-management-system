package auth

import (
	"time"

	"github.com/iSparshP/real-time-task-management-system/internal/models"
)

// Use the models package types
type User = models.User

// Remove the User struct definition and hooks - they're now in models package

// Keep only the request/response types
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type Config struct {
	JWTSecret              string
	TokenExpiration        time.Duration
	RefreshTokenExpiration time.Duration
}
