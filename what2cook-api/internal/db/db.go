package db

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"what2cook-api/internal/auth"
)

// Open opens a SQLite database at path.
func Open(path string) (*gorm.DB, error) {
	if path == "" {
		return nil, fmt.Errorf("db path is empty")
	}

	gdb, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.New(
			log.Default(),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		),
	})
	if err != nil {
		return nil, fmt.Errorf("gorm open: %w", err)
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, fmt.Errorf("sql db: %w", err)
	}
	sqlDB.SetMaxOpenConns(1) // SQLite-friendly

	log.Printf("database opened: %s", path)
	return gdb, nil
}

// AutoMigrate runs GORM migrations for all known models.
func AutoMigrate(gdb *gorm.DB) error {
	if err := gdb.AutoMigrate(
		&auth.User{},
		&auth.Session{},
		&auth.PasswordReset{},
		&auth.EmailVerification{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}
	return nil
}
