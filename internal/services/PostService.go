package services

import (
	"renter_backend/internal/models"

	"gorm.io/gorm"
)

// 只是定義 PostService 是包了一個db的結構 負責文章的商業邏輯
type PostService struct {
	db *gorm.DB
}

// NewPostService 用來建立 PostService 接收 DB 連線
func NewPostService(db *gorm.DB) *PostService {
	return &PostService{db: db}
}

// GetPostForMainPage 取熱門文章
func (s *PostService) GetPostForMainPage(post_filter string) ([]models.Post, error) {
	var posts []models.Post

	// 預設排序：最新
	orderBy := "created_at DESC"

	switch post_filter {
	case "hot":
		orderBy = "like_number DESC"
	case "old":
		orderBy = "created_at ASC"
	case "new":
		orderBy = "created_at DESC"
	}
    //*****  Go 有一個特性：if 可以同時宣告變數 + 判斷條件。  *****//
	if err := s.db.Order(orderBy).Limit(10).Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

// GetPostByID 取單一文章
func (s *PostService) GetPostByID(postID string) (*models.Post, error) {
	var post models.Post
	if err := s.db.First(&post, "post_id = ?", postID).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

