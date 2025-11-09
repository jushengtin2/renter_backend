package models

import (
	"time"
)
type CommentLike struct {

	UserID        string `gorm:"primaryKey type:text;not null" json:"user_id"`
	CommentID int    `gorm:"primaryKey type:int;not null" json:"comment_id"`
	CreatedAt	time.Time `gorm:"autoCreateTime" json:"created_at"`

}