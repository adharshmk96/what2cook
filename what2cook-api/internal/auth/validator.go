package auth

import (
	"strings"
)

// RegisterRequest is the body for POST /auth/register.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest is the body for POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ForgotPasswordRequest is the body for POST /auth/forgot-password.
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest is the body for POST /auth/reset-password.
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// ChangePasswordRequest is the body for POST /auth/change-password.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// ValidateRegister checks register input shape before hitting the service.
func ValidateRegister(req *RegisterRequest) string {
	if strings.TrimSpace(req.Email) == "" {
		return "email is required"
	}
	if req.Password == "" {
		return "password is required"
	}
	if len(req.Password) < minPasswordLen {
		return "password must be at least 8 characters"
	}
	return ""
}

// ValidateLogin checks login input shape.
func ValidateLogin(req *LoginRequest) string {
	if strings.TrimSpace(req.Email) == "" {
		return "email is required"
	}
	if req.Password == "" {
		return "password is required"
	}
	return ""
}

// ValidateForgotPassword checks forgot-password input.
func ValidateForgotPassword(req *ForgotPasswordRequest) string {
	if strings.TrimSpace(req.Email) == "" {
		return "email is required"
	}
	return ""
}

// ValidateResetPassword checks reset-password input.
func ValidateResetPassword(req *ResetPasswordRequest) string {
	if strings.TrimSpace(req.Token) == "" {
		return "token is required"
	}
	if req.NewPassword == "" {
		return "new_password is required"
	}
	if len(req.NewPassword) < minPasswordLen {
		return "new_password must be at least 8 characters"
	}
	return ""
}

// ValidateChangePassword checks change-password input.
func ValidateChangePassword(req *ChangePasswordRequest) string {
	if req.OldPassword == "" {
		return "old_password is required"
	}
	if req.NewPassword == "" {
		return "new_password is required"
	}
	if len(req.NewPassword) < minPasswordLen {
		return "new_password must be at least 8 characters"
	}
	return ""
}
