package models

type User struct {
	UserID        string `gorm:"primaryKey;type:text" json:"user_id"`
	FirstName      string `gorm:"type:text;not null" json:"first_name"`
	LastName     string `gorm:"type:text; not null" json:"last_name"`
	PhoneNumber  string `gorm:"type:text; not null" json:"phone_number"`
	Email        string `gorm:"type:text; not null" json:"email"`
	PasswordHash string `gorm:"type:text; not null" json:"hashed_password"`
	// Picture   string
}