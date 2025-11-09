package models

import (
	"time"

	"gorm.io/datatypes"
)

type Post struct { //他會自動改成snake_case然後給你加s 所以對照到了我sql的table name
	PostID     int       `gorm:"primaryKey;autoIncrement" json:"post_id"`
	UserID     string    `gorm:"type:text;not null" json:"user_id"`
	Title      string    `gorm:"type:varchar(200);not null" json:"title"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	PictureURL datatypes.JSON `gorm:"type:jsonb" json:"picture_url"`
	Location   string    `gorm:"type:varchar(100)" json:"location"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	LikeNumber int       `gorm:"default:0" json:"like_number"`
	SaveNumber int       `gorm:"default:0" json:"save_number"`
	Status     bool      `gorm:"default:true" json:"status"`
	
}


