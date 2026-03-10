package database

import (
	"fmt"
	"log"
	"renter_backend/internal/models"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database is nil")
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Post{},
		&models.Comment{},
	); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	log.Println("migration completed")
	return nil
}
