package auth

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler exposes HTTP handlers for the auth slice.
type Handler struct {
	svc *Service
}

// NewHandler creates an auth handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type errorBody struct {
	Error string `json:"error"`
}

type messageBody struct {
	Message string `json:"message"`
}

type meBody struct {
	User *User `json:"user"`
}

// Register handles POST /auth/register.
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorBody{Error: "invalid JSON body"})
		return
	}
	if msg := ValidateRegister(&req); msg != "" {
		c.JSON(http.StatusBadRequest, errorBody{Error: msg})
		return
	}

	result, err := h.svc.Register(req.Email, req.Password)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

// Login handles POST /auth/login.
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorBody{Error: "invalid JSON body"})
		return
	}
	if msg := ValidateLogin(&req); msg != "" {
		c.JSON(http.StatusBadRequest, errorBody{Error: msg})
		return
	}

	result, err := h.svc.Login(req.Email, req.Password)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// Logout handles POST /auth/logout.
func (h *Handler) Logout(c *gin.Context) {
	sessionID, ok := SessionIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorBody{Error: "unauthorized"})
		return
	}
	if err := h.svc.Logout(sessionID); err != nil {
		log.Printf("logout error: %v", err)
		c.JSON(http.StatusInternalServerError, errorBody{Error: "internal error"})
		return
	}
	c.JSON(http.StatusOK, messageBody{Message: "logged out"})
}

// Me handles GET /auth/me.
func (h *Handler) Me(c *gin.Context) {
	userID, ok := UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorBody{Error: "unauthorized"})
		return
	}
	user, err := h.svc.Me(userID)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, meBody{User: user})
}

// ForgotPassword handles POST /auth/forgot-password.
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorBody{Error: "invalid JSON body"})
		return
	}
	if msg := ValidateForgotPassword(&req); msg != "" {
		c.JSON(http.StatusBadRequest, errorBody{Error: msg})
		return
	}

	if err := h.svc.ForgotPassword(req.Email); err != nil {
		log.Printf("forgot password error: %v", err)
		// Still return generic success to avoid leaking SMTP / account details.
	}
	c.JSON(http.StatusOK, messageBody{Message: "if that email exists, a reset link was sent"})
}

// ResetPassword handles POST /auth/reset-password.
func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorBody{Error: "invalid JSON body"})
		return
	}
	if msg := ValidateResetPassword(&req); msg != "" {
		c.JSON(http.StatusBadRequest, errorBody{Error: msg})
		return
	}

	if err := h.svc.ResetPassword(req.Token, req.NewPassword); err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, messageBody{Message: "password reset"})
}

// ChangePassword handles POST /auth/change-password.
func (h *Handler) ChangePassword(c *gin.Context) {
	userID, ok := UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorBody{Error: "unauthorized"})
		return
	}
	sessionID, ok := SessionIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorBody{Error: "unauthorized"})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorBody{Error: "invalid JSON body"})
		return
	}
	if msg := ValidateChangePassword(&req); msg != "" {
		c.JSON(http.StatusBadRequest, errorBody{Error: msg})
		return
	}

	result, err := h.svc.ChangePassword(userID, sessionID, req.OldPassword, req.NewPassword)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// UpdateMe handles PATCH /auth/me.
func (h *Handler) UpdateMe(c *gin.Context) {
	userID, ok := UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorBody{Error: "unauthorized"})
		return
	}

	var req UpdateEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorBody{Error: "invalid JSON body"})
		return
	}
	if msg := ValidateUpdateEmail(&req); msg != "" {
		c.JSON(http.StatusBadRequest, errorBody{Error: msg})
		return
	}

	user, err := h.svc.UpdateEmail(userID, req.Email)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, meBody{User: user})
}

// VerifyEmail handles POST /auth/verify-email.
func (h *Handler) VerifyEmail(c *gin.Context) {
	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorBody{Error: "invalid JSON body"})
		return
	}
	if msg := ValidateVerifyEmail(&req); msg != "" {
		c.JSON(http.StatusBadRequest, errorBody{Error: msg})
		return
	}

	user, err := h.svc.VerifyEmail(req.Token)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, meBody{User: user})
}

// ResendVerification handles POST /auth/resend-verification.
func (h *Handler) ResendVerification(c *gin.Context) {
	userID, ok := UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorBody{Error: "unauthorized"})
		return
	}

	if err := h.svc.ResendVerification(userID); err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, messageBody{Message: "verification email sent"})
}

func (h *Handler) writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidCredentials), errors.Is(err, ErrWrongPassword):
		c.JSON(http.StatusUnauthorized, errorBody{Error: "invalid credentials"})
	case errors.Is(err, ErrUnauthorized), errors.Is(err, ErrInvalidToken):
		c.JSON(http.StatusUnauthorized, errorBody{Error: "unauthorized"})
	case errors.Is(err, ErrEmailTaken):
		c.JSON(http.StatusConflict, errorBody{Error: "email already registered"})
	case errors.Is(err, ErrWeakPassword), errors.Is(err, ErrInvalidEmail):
		c.JSON(http.StatusBadRequest, errorBody{Error: err.Error()})
	case errors.Is(err, ErrInvalidResetToken):
		c.JSON(http.StatusBadRequest, errorBody{Error: "invalid or expired reset token"})
	case errors.Is(err, ErrInvalidVerifyToken):
		c.JSON(http.StatusBadRequest, errorBody{Error: "invalid or expired verification token"})
	case errors.Is(err, ErrEmailAlreadyVerified):
		c.JSON(http.StatusBadRequest, errorBody{Error: "email already verified"})
	default:
		log.Printf("auth error: %v", err)
		c.JSON(http.StatusInternalServerError, errorBody{Error: "internal error"})
	}
}
