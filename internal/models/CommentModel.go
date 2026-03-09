package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm" // 記得引入這個
)

type Comment struct {
	CommentID      int            `gorm:"primaryKey;autoIncrement" json:"comment_id"`
	Content        string         `gorm:"type:text;not null" json:"content"`
	CreatedAt      time.Time      `json:"created_at"`
	UserID         string         `gorm:"type:text;not null" json:"user_id"`
	PictureURL     datatypes.JSON `gorm:"type:jsonb" json:"picture_url"` // 支援多張圖片
	ReplyPostID    *int           `gorm:"column:reply_post_id" json:"reply_post_id"`       // 對應哪一篇貼文
	ReplyCommentID *int           `gorm:"column:reply_comment_id" json:"reply_comment_id"` // 回覆哪一則留言（若有）
	LikeNumber    int      		  `gorm:"default:0" json:"like_number"`
	User          User           `gorm:"-:migration;foreignKey:UserID;references:UserID"` // 保留 Preload 關聯，但不由 AutoMigrate 建 FK
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"` // 軟刪除核心：這行讓查詢自動排除已刪除貼文
	// ILikeItOrNot  bool 			  `gorm:"<-:false" json:"i_like_it_or_not"` //`gorm:"<-:false" 代表只讀不寫
}


