package services

import (
	"fmt"
	"renter_backend/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

type UserService struct {
	db *gorm.DB
}

func (us *UserService) GetUserProfile(genre string, userID string, firstName string, lastName string, email string, pictureURL string) (interface{}, error) {
	switch genre {
	case "created":
		user := models.User{
			UserID:         userID,
			FirstName:      firstName,
			LastName:       lastName,
			Email:          email,
			ProfilePicture: pictureURL,
		}

		if err := us.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"first_name", "last_name", "email", "profile_picture"}),
		}).Create(&user).Error; err != nil {
			return nil, err
		}
		return &user, nil
	case "updated":
		updates := map[string]interface{}{
			"first_name":      firstName,
			"last_name":       lastName,
			"profile_picture": pictureURL,
		}
		if email != "" {
			updates["email"] = email
		}

		if err := us.db.Model(&models.User{}).
			Where("user_id = ?", userID).
			Updates(updates).Error; err != nil {
			return nil, err
		}
		return nil, nil
	case "deleted":
		if err := us.db.Delete(&models.User{}, "user_id = ?", userID).Error; err != nil {
			return nil, err
		}
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown user event genre: %s", genre)
	}
}
