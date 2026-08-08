package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrNotFound = errors.New("not found")
)

// Repository persists auth entities.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates an auth repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// CreateUser inserts a user.
func (r *Repository) CreateUser(user *User) error {
	if err := r.db.Create(user).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

// FindUserByEmail looks up a user by email (case-sensitive as stored).
func (r *Repository) FindUserByEmail(email string) (*User, error) {
	var user User
	err := r.db.Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return &user, nil
}

// FindUserByID looks up a user by id.
func (r *Repository) FindUserByID(id uuid.UUID) (*User, error) {
	var user User
	err := r.db.Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &user, nil
}

// UpdateUserPassword sets the password hash for a user.
func (r *Repository) UpdateUserPassword(userID uuid.UUID, passwordHash string) error {
	res := r.db.Model(&User{}).Where("id = ?", userID).Update("password_hash", passwordHash)
	if res.Error != nil {
		return fmt.Errorf("update password: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// CreateSession inserts a session.
func (r *Repository) CreateSession(session *Session) error {
	if err := r.db.Create(session).Error; err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

// FindSessionByID loads a session by id.
func (r *Repository) FindSessionByID(id uuid.UUID) (*Session, error) {
	var session Session
	err := r.db.Where("id = ?", id).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find session: %w", err)
	}
	return &session, nil
}

// ExtendSessionExpiry updates expires_at (sliding refresh).
func (r *Repository) ExtendSessionExpiry(id uuid.UUID, expiresAt time.Time) error {
	res := r.db.Model(&Session{}).Where("id = ?", id).Update("expires_at", expiresAt)
	if res.Error != nil {
		return fmt.Errorf("extend session: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteSession removes a session by id.
func (r *Repository) DeleteSession(id uuid.UUID) error {
	res := r.db.Where("id = ?", id).Delete(&Session{})
	if res.Error != nil {
		return fmt.Errorf("delete session: %w", res.Error)
	}
	return nil
}

// DeleteSessionsByUserID removes all sessions for a user.
func (r *Repository) DeleteSessionsByUserID(userID uuid.UUID) error {
	if err := r.db.Where("user_id = ?", userID).Delete(&Session{}).Error; err != nil {
		return fmt.Errorf("delete sessions by user: %w", err)
	}
	return nil
}

// CreatePasswordReset inserts a reset record.
func (r *Repository) CreatePasswordReset(reset *PasswordReset) error {
	if err := r.db.Create(reset).Error; err != nil {
		return fmt.Errorf("create password reset: %w", err)
	}
	return nil
}

// FindValidPasswordReset finds an unused, unexpired reset by token hash.
func (r *Repository) FindValidPasswordReset(tokenHash string, now time.Time) (*PasswordReset, error) {
	var reset PasswordReset
	err := r.db.
		Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", tokenHash, now).
		First(&reset).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find password reset: %w", err)
	}
	return &reset, nil
}

// MarkPasswordResetUsed sets used_at.
func (r *Repository) MarkPasswordResetUsed(id uuid.UUID, usedAt time.Time) error {
	res := r.db.Model(&PasswordReset{}).Where("id = ?", id).Update("used_at", usedAt)
	if res.Error != nil {
		return fmt.Errorf("mark reset used: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateUserEmail sets email and clears email verification.
func (r *Repository) UpdateUserEmail(userID uuid.UUID, email string) error {
	res := r.db.Model(&User{}).Where("id = ?", userID).Updates(map[string]any{
		"email":             email,
		"email_verified_at": nil,
	})
	if res.Error != nil {
		return fmt.Errorf("update email: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// MarkUserEmailVerified sets email_verified_at.
func (r *Repository) MarkUserEmailVerified(userID uuid.UUID, verifiedAt time.Time) error {
	res := r.db.Model(&User{}).Where("id = ?", userID).Update("email_verified_at", verifiedAt)
	if res.Error != nil {
		return fmt.Errorf("mark email verified: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// CreateEmailVerification inserts a verification record.
func (r *Repository) CreateEmailVerification(v *EmailVerification) error {
	if err := r.db.Create(v).Error; err != nil {
		return fmt.Errorf("create email verification: %w", err)
	}
	return nil
}

// FindValidEmailVerification finds an unused, unexpired verification by token hash.
func (r *Repository) FindValidEmailVerification(tokenHash string, now time.Time) (*EmailVerification, error) {
	var v EmailVerification
	err := r.db.
		Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", tokenHash, now).
		First(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find email verification: %w", err)
	}
	return &v, nil
}

// MarkEmailVerificationUsed sets used_at.
func (r *Repository) MarkEmailVerificationUsed(id uuid.UUID, usedAt time.Time) error {
	res := r.db.Model(&EmailVerification{}).Where("id = ?", id).Update("used_at", usedAt)
	if res.Error != nil {
		return fmt.Errorf("mark verification used: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
