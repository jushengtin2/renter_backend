package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm" // 記得引入這個
)

type Post struct {
    PostID     int            `gorm:"primaryKey;autoIncrement" json:"post_id"`
    UserID     string         `gorm:"type:text;not null;index" json:"user_id"` // 建議加 index 提升查詢效率
    Title      string         `gorm:"type:varchar(200);not null" json:"title"`
    Content    string         `gorm:"type:text;not null" json:"content"`
    Picture    datatypes.JSON `gorm:"type:jsonb" json:"picture"`
    Address    string         `gorm:"type:varchar(100)" json:"address"`
    City       string         `gorm:"type:varchar(50)" json:"city"`
    District   string         `gorm:"type:varchar(50)" json:"district"`
    Latitude   float64        `gorm:"type:decimal(10,6)" json:"latitude"`
    Longitude  float64        `gorm:"type:decimal(10,6)" json:"longitude"`
    Location   []byte         `gorm:"type:geography(Point,4326)" json:"location,omitempty"`
    CreatedAt  time.Time      `json:"created_at"`
    UpdatedAt  time.Time      `json:"updated_at"` // 建議增加更新時間
    DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"` // 軟刪除核心：這行讓查詢自動排除已刪除貼文
    LikeNumber int            `gorm:"default:0" json:"like_number"`
    SaveNumber int            `gorm:"default:0" json:"save_number"`
    Tags       datatypes.JSON `gorm:"type:jsonb" json:"tags"`
}