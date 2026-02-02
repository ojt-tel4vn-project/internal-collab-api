package auth

import "github.com/google/uuid"

// Login DTOs
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken           string `json:"access_token"`
	TokenType             string `json:"token_type"`
	RequirePasswordChange bool   `json:"require_password_change"` // True if first-time login
	User                  struct {
		ID           uuid.UUID `json:"id"`
		Email        string    `json:"email"`
		Name         string    `json:"name"`
		EmployeeCode string    `json:"employee_code"`
		Status       string    `json:"status"`
	} `json:"user"`
}

// Change Password DTOs
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type ChangePasswordResponse struct {
	Message string `json:"message"`
}

// First-time Setup DTOs (for employees created by HR)
type FirstTimeSetupRequest struct {
	TemporaryPassword string `json:"temporary_password" binding:"required"`
	NewPassword       string `json:"new_password" binding:"required,min=8"`
}

type FirstTimeSetupResponse struct {
	Message     string `json:"message"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}
