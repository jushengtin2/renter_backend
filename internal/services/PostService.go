package services

import (
	"fmt"
	"renter_backend/internal/models"
	"time"

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
func (ps *PostService) GetPostForMainPage(post_filter string) ([]models.Post, error) {
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
	if err := ps.db.Order(orderBy).Limit(10).Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

// GetPostByID 取單一文章
func (ps *PostService) GetPostByID(postID string) (*models.Post, error) {
	var post models.Post
	if err := ps.db.First(&post, "post_id = ?", postID).Error; err != nil {
		return nil, err
	}
	fmt.Println(&post)
	return &post, nil
}

func (s *PostService) CreatePost(
    userID string,
    title string,
    picture string,
    location string,
    // coordinate string,
    //time不用傳我在service層直接寫time.now()
    content string,
    tags []string,
) (*models.Post, error) {
    post := models.Post{
        UserID:      userID,
        Title:       title,
        PictureURL:  picture,
        Location:    location,
        CreatedAt:    time.Now(),
        Content:     content,
        // Tags:        strings,  要寫在另外一個table (tags<->posts)
    }

    if err := s.db.Create(&post).Error; err != nil {
        return nil, err
    }
    return &post, nil
}


func (s *PostService) DeletePost(postID int) error {
	if err := s.db.Delete(&models.Post{}, "post_id = ?", postID).Error; err != nil {
		return err
	}
	return nil
}


