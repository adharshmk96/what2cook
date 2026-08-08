package auth

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User is an application account.
type User struct {
	ID              uuid.UUID  `gorm:"type:text;primaryKey" json:"id"`
	Email           string     `gorm:"uniqueIndex;size:255;not null" json:"email"`
	PasswordHash    string     `gorm:"not null" json:"-"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// EmailVerified reports whether the user has verified their email.
func (u *User) EmailVerified() bool {
	return u != nil && u.EmailVerifiedAt != nil
}

// BeforeCreate assigns a UUID if missing.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// Session is a server-side login session validated by opaque tokens.
type Session struct {
	ID        uuid.UUID `gorm:"type:text;primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:text;index;not null" json:"user_id"`
	ExpiresAt time.Time `gorm:"index;not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate assigns a UUID if missing.
func (s *Session) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// PasswordReset stores a hashed one-time reset token.
type PasswordReset struct {
	ID        uuid.UUID  `gorm:"type:text;primaryKey"`
	UserID    uuid.UUID  `gorm:"type:text;index;not null"`
	TokenHash string     `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time  `gorm:"index;not null"`
	UsedAt    *time.Time `gorm:"index"`
	CreatedAt time.Time
}

// BeforeCreate assigns a UUID if missing.
func (p *PasswordReset) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// EmailVerification stores a hashed one-time email verification token.
type EmailVerification struct {
	ID        uuid.UUID  `gorm:"type:text;primaryKey"`
	UserID    uuid.UUID  `gorm:"type:text;index;not null"`
	TokenHash string     `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time  `gorm:"index;not null"`
	UsedAt    *time.Time `gorm:"index"`
	CreatedAt time.Time
}

// BeforeCreate assigns a UUID if missing.
func (e *EmailVerification) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}
