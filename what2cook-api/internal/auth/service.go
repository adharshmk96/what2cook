package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	minPasswordLen = 8
	bcryptCost     = bcrypt.DefaultCost
)

var (
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrInvalidToken         = errors.New("invalid token")
	ErrEmailTaken           = errors.New("email already registered")
	ErrWeakPassword         = errors.New("password too weak")
	ErrInvalidEmail         = errors.New("invalid email")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrWrongPassword        = errors.New("wrong password")
	ErrInvalidResetToken    = errors.New("invalid or expired reset token")
	ErrInvalidVerifyToken   = errors.New("invalid or expired verification token")
	ErrEmailAlreadyVerified = errors.New("email already verified")
)

// tokenPayload is the unsigned JSON body of an opaque auth token.
type tokenPayload struct {
	UserID    string `json:"userId"`
	SessionID string `json:"sessionId"`
}

// AuthMailer sends password-reset and email-verification links.
type AuthMailer interface {
	SendPasswordReset(toEmail, rawToken string) error
	SendEmailVerification(toEmail, rawToken string) error
}

// Service contains auth business logic.
type Service struct {
	repo        *Repository
	mailer      AuthMailer
	tokenSecret []byte
	sessionTTL  time.Duration
	resetTTL    time.Duration
	verifyTTL   time.Duration
}

// NewService creates an auth service.
func NewService(repo *Repository, mailer AuthMailer, tokenSecret string, sessionTTL, resetTTL, verifyTTL time.Duration) *Service {
	return &Service{
		repo:        repo,
		mailer:      mailer,
		tokenSecret: []byte(tokenSecret),
		sessionTTL:  sessionTTL,
		resetTTL:    resetTTL,
		verifyTTL:   verifyTTL,
	}
}

// AuthResult is returned after register/login.
type AuthResult struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

// Register creates a user, sends a verification email (best-effort), and issues a session.
func (s *Service) Register(email, password string) (*AuthResult, error) {
	email, err := normalizeEmail(email)
	if err != nil {
		return nil, err
	}
	if err := validatePassword(password); err != nil {
		return nil, err
	}

	if _, err := s.repo.FindUserByEmail(email); err == nil {
		return nil, ErrEmailTaken
	} else if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &User{Email: email, PasswordHash: string(hash)}
	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}

	if err := s.sendVerificationEmail(user); err != nil {
		log.Printf("send verification after register for %s: %v", user.Email, err)
	}

	return s.issueSession(user)
}

// Login verifies credentials and creates a session.
func (s *Service) Login(email, password string) (*AuthResult, error) {
	email, err := normalizeEmail(email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := s.repo.FindUserByEmail(email)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.issueSession(user)
}

// Logout deletes the session for the given session id.
func (s *Service) Logout(sessionID uuid.UUID) error {
	return s.repo.DeleteSession(sessionID)
}

// Me returns the current user.
func (s *Service) Me(userID uuid.UUID) (*User, error) {
	user, err := s.repo.FindUserByID(userID)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrUnauthorized
	}
	return user, err
}

// UpdateEmail changes the user's email, clears verification, and sends a new verify link.
func (s *Service) UpdateEmail(userID uuid.UUID, newEmail string) (*User, error) {
	email, err := normalizeEmail(newEmail)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.FindUserByID(userID)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrUnauthorized
	}
	if err != nil {
		return nil, err
	}

	if email == user.Email {
		return user, nil
	}

	if existing, err := s.repo.FindUserByEmail(email); err == nil && existing.ID != userID {
		return nil, ErrEmailTaken
	} else if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	if err := s.repo.UpdateUserEmail(userID, email); err != nil {
		return nil, err
	}

	user.Email = email
	user.EmailVerifiedAt = nil

	if err := s.sendVerificationEmail(user); err != nil {
		log.Printf("send verification after email update for %s: %v", user.Email, err)
	}

	return user, nil
}

// VerifyEmail consumes a verification token and marks the user's email verified.
func (s *Service) VerifyEmail(rawToken string) (*User, error) {
	if strings.TrimSpace(rawToken) == "" {
		return nil, ErrInvalidVerifyToken
	}

	tokenHash := hashToken(rawToken)
	v, err := s.repo.FindValidEmailVerification(tokenHash, time.Now().UTC())
	if errors.Is(err, ErrNotFound) {
		return nil, ErrInvalidVerifyToken
	}
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if err := s.repo.MarkUserEmailVerified(v.UserID, now); err != nil {
		return nil, err
	}
	if err := s.repo.MarkEmailVerificationUsed(v.ID, now); err != nil {
		return nil, err
	}

	user, err := s.repo.FindUserByID(v.UserID)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrInvalidVerifyToken
	}
	return user, err
}

// ResendVerification sends a new verification email when the account is unverified.
func (s *Service) ResendVerification(userID uuid.UUID) error {
	user, err := s.repo.FindUserByID(userID)
	if errors.Is(err, ErrNotFound) {
		return ErrUnauthorized
	}
	if err != nil {
		return err
	}
	if user.EmailVerified() {
		return ErrEmailAlreadyVerified
	}
	return s.sendVerificationEmail(user)
}

func (s *Service) sendVerificationEmail(user *User) error {
	rawToken, err := randomToken(32)
	if err != nil {
		return err
	}
	v := &EmailVerification{
		UserID:    user.ID,
		TokenHash: hashToken(rawToken),
		ExpiresAt: time.Now().UTC().Add(s.verifyTTL),
	}
	if err := s.repo.CreateEmailVerification(v); err != nil {
		return err
	}
	if err := s.mailer.SendEmailVerification(user.Email, rawToken); err != nil {
		return fmt.Errorf("send verification mail: %w", err)
	}
	return nil
}

// ForgotPassword creates a reset token and emails/logs the link. Always succeeds for unknown emails (no enumeration).
func (s *Service) ForgotPassword(email string) error {
	emailNorm, err := normalizeEmail(email)
	if err != nil {
		// Still return success to avoid leaking validation differences.
		return nil
	}

	user, err := s.repo.FindUserByEmail(emailNorm)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	rawToken, err := randomToken(32)
	if err != nil {
		return err
	}
	tokenHash := hashToken(rawToken)

	reset := &PasswordReset{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().UTC().Add(s.resetTTL),
	}
	if err := s.repo.CreatePasswordReset(reset); err != nil {
		return err
	}

	if err := s.mailer.SendPasswordReset(user.Email, rawToken); err != nil {
		return fmt.Errorf("send reset mail: %w", err)
	}
	return nil
}

// ResetPassword consumes a reset token, sets a new password, and invalidates sessions.
func (s *Service) ResetPassword(rawToken, newPassword string) error {
	if strings.TrimSpace(rawToken) == "" {
		return ErrInvalidResetToken
	}
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	tokenHash := hashToken(rawToken)
	reset, err := s.repo.FindValidPasswordReset(tokenHash, time.Now().UTC())
	if errors.Is(err, ErrNotFound) {
		return ErrInvalidResetToken
	}
	if err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := s.repo.UpdateUserPassword(reset.UserID, string(hash)); err != nil {
		return err
	}
	now := time.Now().UTC()
	if err := s.repo.MarkPasswordResetUsed(reset.ID, now); err != nil {
		return err
	}
	return s.repo.DeleteSessionsByUserID(reset.UserID)
}

// ChangePassword verifies the old password, sets a new one, and rotates the session.
func (s *Service) ChangePassword(userID, sessionID uuid.UUID, oldPassword, newPassword string) (*AuthResult, error) {
	if err := validatePassword(newPassword); err != nil {
		return nil, err
	}

	user, err := s.repo.FindUserByID(userID)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrUnauthorized
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return nil, ErrWrongPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	if err := s.repo.UpdateUserPassword(user.ID, string(hash)); err != nil {
		return nil, err
	}

	// Rotate: drop current session and issue a new one.
	_ = s.repo.DeleteSession(sessionID)
	return s.issueSession(user)
}

// ValidateToken parses and verifies an opaque HMAC token, loads the session, and applies sliding refresh.
func (s *Service) ValidateToken(token string) (userID, sessionID uuid.UUID, err error) {
	payload, err := s.parseAndVerifyToken(token)
	if err != nil {
		return uuid.Nil, uuid.Nil, ErrInvalidToken
	}

	uid, err := uuid.Parse(payload.UserID)
	if err != nil {
		return uuid.Nil, uuid.Nil, ErrInvalidToken
	}
	sid, err := uuid.Parse(payload.SessionID)
	if err != nil {
		return uuid.Nil, uuid.Nil, ErrInvalidToken
	}

	session, err := s.repo.FindSessionByID(sid)
	if errors.Is(err, ErrNotFound) {
		return uuid.Nil, uuid.Nil, ErrUnauthorized
	}
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	now := time.Now().UTC()
	if session.UserID != uid || !session.ExpiresAt.After(now) {
		return uuid.Nil, uuid.Nil, ErrUnauthorized
	}

	newExpiry := now.Add(s.sessionTTL)
	if err := s.repo.ExtendSessionExpiry(sid, newExpiry); err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	return uid, sid, nil
}

func (s *Service) issueSession(user *User) (*AuthResult, error) {
	session := &Session{
		UserID:    user.ID,
		ExpiresAt: time.Now().UTC().Add(s.sessionTTL),
	}
	if err := s.repo.CreateSession(session); err != nil {
		return nil, err
	}

	token, err := s.signToken(user.ID, session.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{Token: token, User: user}, nil
}

func (s *Service) signToken(userID, sessionID uuid.UUID) (string, error) {
	payload := tokenPayload{
		UserID:    userID.String(),
		SessionID: sessionID.String(),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal token: %w", err)
	}
	payloadB64 := base64.RawURLEncoding.EncodeToString(raw)
	sig := s.hmacSign(payloadB64)
	return payloadB64 + "." + sig, nil
}

func (s *Service) parseAndVerifyToken(token string) (*tokenPayload, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, ErrInvalidToken
	}
	expected := s.hmacSign(parts[0])
	if !hmac.Equal([]byte(expected), []byte(parts[1])) {
		return nil, ErrInvalidToken
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var payload tokenPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, ErrInvalidToken
	}
	if payload.UserID == "" || payload.SessionID == "" {
		return nil, ErrInvalidToken
	}
	return &payload, nil
}

func (s *Service) hmacSign(payloadB64 string) string {
	mac := hmac.New(sha256.New, s.tokenSecret)
	_, _ = mac.Write([]byte(payloadB64))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("random token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func normalizeEmail(email string) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || !strings.Contains(email, "@") || strings.Contains(email, " ") {
		return "", ErrInvalidEmail
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || !strings.Contains(parts[1], ".") {
		return "", ErrInvalidEmail
	}
	return email, nil
}

func validatePassword(password string) error {
	if len(password) < minPasswordLen {
		return ErrWeakPassword
	}
	return nil
}
