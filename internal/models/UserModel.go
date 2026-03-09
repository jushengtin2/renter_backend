package models

type User struct {
	UserID        string `gorm:"primaryKey;type:text" json:"user_id"`
	FirstName      string `gorm:"type:text;not null" json:"first_name"`
	LastName     string `gorm:"type:text; not null" json:"last_name"`
	Email        string `gorm:"type:text;not null" json:"email"`
	ProfilePicture string `gorm:"type:text" json:"profile_picture"`

}
