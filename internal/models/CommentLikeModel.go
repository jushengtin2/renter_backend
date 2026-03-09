package models

import (
	"time"
)
type CommentLike struct {

	UserID        string `gorm:"primaryKey type:text;not null" json:"user_id"`
	CommentID int    `gorm:"primaryKey type:int;not null" json:"comment_id"`
	CreatedAt	time.Time `json:"created_at"`

}