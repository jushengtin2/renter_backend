package database

import (
	"log"
	"renter_backend/internal/models"
)

func AutoMigrate() {
	if err := DB.AutoMigrate(&models.Post{}); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
}
