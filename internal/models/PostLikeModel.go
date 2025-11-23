package models

import (
	"time"
)
type PostLike struct {

	UserID        string `gorm:"primaryKey type:text;not null" json:"user_id"`
	PostID int    `gorm:"primaryKey type:int;not null" json:"post_id"`
	CreatedAt	time.Time `gorm:"autoCreateTime" json:"created_at"`

}