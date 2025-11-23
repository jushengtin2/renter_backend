package services

import (
	"encoding/json"
	"fmt"
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

type MainPagePostResponse struct {
	PostID         int      `json:"post_id"`
	Title          string   `json:"title"`
	Timestamp      string   `json:"timestamp"`
	Content        string   `json:"content"`
	LikeNumber    int      `json:"like_number"`
	BookmarkNumber int      `json:"bookmark_number"`
	Tags           []string `json:"tags"`
	ILikeThis      bool     `json:"i_like_this"`
	ISaveThis      bool     `json:"i_save_this"`
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

// GetPostForMainPage 取熱門文章
func (ps *PostService) GetPostForMainPage(post_filter string, userID string) ([]*MainPagePostResponse, error) {
    var posts []models.Post

    // 排序方式
    orderBy := "created_at DESC"
    switch post_filter {
    case "hot":
        orderBy = "like_number DESC"
    case "old":
        orderBy = "created_at ASC"
    case "new":
        orderBy = "created_at DESC"
    }

    // 撈出 10 篇貼文
    if err := ps.db.Order(orderBy).Limit(10).Find(&posts).Error; err != nil {
        return nil, err
    }

    if len(posts) == 0 {
        return []*MainPagePostResponse{}, nil
    }

    // 整理 postIDs
    postIDs := make([]int, len(posts))
    for i, p := range posts {
        postIDs[i] = p.PostID
    }

    // -------------------------
    //  查詢使用者是否按讚
    // -------------------------
    likedMap := make(map[int]bool)
    if userID != "" {
        var likedIDs []int
        ps.db.Table("post_likes").
            Select("post_id").
            Where("user_id = ? AND post_id IN ?", userID, postIDs).
            Pluck("post_id", &likedIDs)

        for _, id := range likedIDs {
            likedMap[id] = true
        }
    }

    // -------------------------
    //  查詢是否收藏
    // -------------------------
    savedMap := make(map[int]bool)
    if userID != "" {
        var saveIDs []int
        ps.db.Table("post_saves").
            Select("post_id").
            Where("user_id = ? AND post_id IN ?", userID, postIDs).
            Pluck("post_id", &saveIDs)

        for _, id := range saveIDs {
            savedMap[id] = true
        }
    }

    // -------------------------
    // 組合 response DTO
    // -------------------------
    res := make([]*MainPagePostResponse, 0, len(posts))

    for _, p := range posts {

        // 轉換 JSONB pictureURL → []string（若需要）
        var pics []string
        if len(p.PictureURL) > 0 {
            json.Unmarshal(p.PictureURL, &pics)
        }

        res = append(res, &MainPagePostResponse{
            PostID:         p.PostID,
            Title:          p.Title,
            Timestamp:      p.CreatedAt.String(),
            Content:        p.Content,
            LikeNumber:     p.LikeNumber,
            BookmarkNumber: p.SaveNumber,
            Tags:           nil,                    // 目前不用 tags
            ILikeThis:      likedMap[p.PostID],     // O(1)
            ISaveThis:      savedMap[p.PostID],     // O(1)
        })
    }

    return res, nil
}


func (ps *PostService) GetPostByID(postID string, userID string) (*PostResponse, error) {
	var post models.Post
	var tagNames []string
	var ilikethis bool
	var isavethis bool

	// 撈文章
	if err := ps.db.First(&post, "post_id = ?", postID).Error; err != nil {
		return nil, err
	}

	// JOIN 把 tags 名稱撈出來(這感覺可以用快取來優化速度)
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
	fmt.Println("startttt")
	fmt.Println(userID)

	// user 已登入
	if userID != "" {
		// 是否按讚
		{
			var exists bool
			err := ps.db.Model(&models.PostLike{}).
				Select("1").
				Where("user_id = ? AND post_id = ?", userID, postID).
				Limit(1).
				Scan(&exists).Error

			if err != nil {
				return nil, err
			}

			ilikethis = exists
			fmt.Println("yes or no:", ilikethis)

		}

		// 是否收藏
		{
			var exists bool
			err := ps.db.Model(&models.PostSave{}).
				Select("1").
				Where("user_id = ? AND post_id = ?", userID, postID).
				Limit(1).
				Scan(&exists).Error

			if err != nil {
				return nil, err
			}

			isavethis = exists
		}

	} else {
		ilikethis = false
		isavethis = false
	}

	//這邊會去看使用者有沒有對這篇文章按讚或收藏 然後直接賦值給下面的DTO
	

	// 組合 DTO
	res := &PostResponse{
		PostID:         post.PostID,
		PictureURL:     pics, // TODO: 改成從 DB 拿
		Author:         "南宜",                                 // TODO: JOIN user 表拿
		AuthorPic:      "https://github.com/shadcn.png",        // TODO: JOIN user 表拿
		Title:          post.Title,
		Timestamp:      post.CreatedAt.String(),
		Content:        post.Content,
		LikeNumber:    	post.LikeNumber,
		BookmarkNumber: post.SaveNumber,
		Tags:           tagNames, 
		ILikeThis:      ilikethis,    
		ISaveThis:      isavethis,   
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

