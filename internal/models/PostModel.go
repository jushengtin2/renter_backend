package models

import (
	"time"
)

type Post struct {
	PostID     int       `gorm:"primaryKey;autoIncrement" json:"post_id"`
	UserID     string    `gorm:"type:uuid;not null" json:"user_id"`
	Title      string    `gorm:"type:varchar(200);not null" json:"title"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	PictureURL string    `gorm:"type:text" json:"picture_url"`
	Location   string    `gorm:"type:varchar(100)" json:"location"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	LikeNumber int       `gorm:"default:0" json:"like_number"`
	SaveNumber int       `gorm:"default:0" json:"save_number"`
	Status     bool      `gorm:"default:true" json:"status"`
}
