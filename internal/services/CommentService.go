package services

import (
	"context"
	"encoding/json"
	"fmt"
	"renter_backend/internal/models"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type CommentService struct {
	db *gorm.DB
	rdb *redis.Client
}

func NewCommentService(db *gorm.DB, rdb *redis.Client) *CommentService {
	return &CommentService{db: db, rdb: rdb}
}

type CommentResponse struct {
	CommentID        int               	`json:"comment_id"`
	UserID 		string            		`json:"user_id"`
	UserFirstName string            	`json:"first_name"`
	UserLastName  string            	`json:"last_name"`
	// UserPicture string            	`json:"user_picture"`
	CommentContent   string            	`json:"content"`
	CommentTime      string            	`json:"created_at"`
	CommentPicture   []string          	`json:"picture_url"`
	CommentLikeNumber int               `json:"like_number"`
	CommentCount int 			 		`json:"comment_number"` 
	ILikeItOrNot   	bool             	`json:"i_like_it_or_not"`
	CommentChild     []CommentResponse `json:"comment_child"`
	ReplyCommentID  *int              	`json:"reply_comment_id,omitempty"`
}


func (cs *CommentService) GetCommentsByPostID(postID string, userID string) ([]*CommentResponse, error) {
	var comments []models.Comment
	fmt.Println("Fetching comments for postID:", postID, "and userID:", userID)

	// 從資料庫撈出指定貼文的留言資料
	if err := cs.db.
		Preload("User").
		Where("reply_post_id = ?", postID).
		Find(&comments).Error; err != nil {
		return nil, err
	}
	key := userID
	ctx := context.Background() //在這裡用 background context 代表「沒有特殊中斷條件」。
	
	commentMap := make(map[int]*CommentResponse)
	var roots []*CommentResponse

	for _, c := range comments {
		var pics []string
		if len(c.PictureURL) > 0 {
			if err := json.Unmarshal(c.PictureURL, &pics); err != nil {
				return nil, err
			}
		}

		// 檢查 Redis 中該留言是否被此使用者按過讚
		isLiked := false
		if userID != "" {
			ok, err := cs.rdb.SIsMember(ctx, key, c.CommentID).Result() 
			if err != nil {
				fmt.Println("Redis 查詢失敗:", err)
			} else {
				isLiked = ok
			}
		}

		commentMap[c.CommentID] = &CommentResponse{
			UserID:           c.UserID,
			UserFirstName:    c.User.FirstName,
			UserLastName:     c.User.LastName,
			CommentID:        c.CommentID,
			CommentContent:   c.Content,
			CommentTime:      c.CreatedAt.String(),
			CommentPicture:   pics,
			CommentCount:     0,
			CommentLikeNumber: c.LikeNumber,
			ILikeItOrNot:     isLiked,
			ReplyCommentID:   c.ReplyCommentID,
			CommentChild:     []CommentResponse{},
		}
	}

	for _, c := range comments {
		if c.ReplyCommentID != nil {
			if parent, ok := commentMap[*c.ReplyCommentID]; ok {
				parent.CommentChild = append(parent.CommentChild, *commentMap[c.CommentID])
				parent.CommentCount++
			}
		} else {
			roots = append(roots, commentMap[c.CommentID])
		}
	}

	if len(roots) > 0 {
		fmt.Println("Response sample:", *roots[0])
	}
	return roots, nil
}
