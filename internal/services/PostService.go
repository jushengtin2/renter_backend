package services

import (
	"encoding/json"
	"renter_backend/internal/models"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// 只是定義 PostService 是包了一個db的結構 負責文章的商業邏輯
type PostService struct {
	db *gorm.DB
	rdb *redis.Client
}

// NewPostService 用來建立 PostService 接收 DB 連線
func NewPostService(db *gorm.DB, rdb *redis.Client) *PostService {
	return &PostService{db: db, rdb: rdb}
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
//不只要去post table抓 還要去post_tags, tags table抓
type PostResponse struct {
	PostID         int      `json:"post_id"`
	PictureURL     []string `json:"picture_url"`
	Author         string   `json:"author"`
	AuthorPic      string   `json:"author_pic"`
	Title          string   `json:"title"`
	Timestamp      string   `json:"timestamp"`
	Content        string   `json:"content"`
	LikeNumber    int      `json:"like_number"`
	BookmarkNumber int      `json:"bookmark_number"`
	Tags           []string `json:"tags"`
	ILikeThis      bool     `json:"i_like_this"`
	ISaveThis      bool     `json:"i_save_this"`
}

func (ps *PostService) GetPostByID(postID string) (*PostResponse, error) {
	var post models.Post
	var tagNames []string

	// 撈文章
	if err := ps.db.First(&post, "post_id = ?", postID).Error; err != nil {
		return nil, err
	}
	

	// JOIN 把 tags 名稱撈出來
	if err := ps.db.Table("tags").
		Select("tags.tag_name").
		Joins("JOIN post_tags ON post_tags.tag_id = tags.tag_id").
		Where("post_tags.post_id = ?", postID).
		Pluck("tag_name", &tagNames).Error; err != nil {
		return nil, err
	}
	var pics []string //因為在postgre存的是jsonb格式 但go沒這種格式 所以要轉成string array
	if len(post.PictureURL) > 0 {
		if err := json.Unmarshal(post.PictureURL, &pics); err != nil {
			return nil, err
		}
	}

	//這邊會抓快取（去看使用者有沒有對這篇文章按讚或收藏）然後直接賦值給下面的DTO
	

	// 組合 DTO
	res := &PostResponse{
		PostID:         post.PostID,
		PictureURL:     pics, // TODO: 改成從 DB 拿
		Author:         "南宜",                                 // TODO: JOIN user 表拿
		AuthorPic:      "https://github.com/shadcn.png",        // TODO: JOIN user 表拿
		Title:          post.Title,
		Timestamp:      post.CreatedAt.String(),
		Content:        post.Content,
		LikeNumber:    post.LikeNumber,
		BookmarkNumber: post.SaveNumber,
		Tags:           tagNames, 
		ILikeThis:      false,    
		ISaveThis:      false,   
	}

	return res, nil
}



func (s *PostService) CreatePost(
    userID string,
    title string,
    picture datatypes.JSON,
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


