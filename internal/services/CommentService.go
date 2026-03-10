package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"renter_backend/internal/models"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CommentService struct {
	db *gorm.DB
	//rdb *redis.Client
	gcsClient *storage.Client
}

func NewCommentService(db *gorm.DB, gcsClient *storage.Client) *CommentService {
	return &CommentService{db: db, gcsClient: gcsClient}
}

type CommentResponse struct {
	CommentID        int               	`json:"comment_id"`
	UserID 		string            		`json:"user_id"`
	UserFirstName string            	`json:"first_name"`
	UserLastName  string            	`json:"last_name"`
	ProfilePicture string            	`json:"profile_picture"`
	ReplyPostID	int               		`json:"reply_post_id"`
	CommentContent   string            	`json:"content"`
	CommentTime      string            	`json:"created_at"`
	CommentPicture   []string          	`json:"picture_url"`
	CommentLikeNumber int               `json:"like_number"`
	CommentCount int 			 		`json:"comment_number"` 
	ILikeItOrNot   	bool             	`json:"i_like_this"`
	CommentChild     []CommentResponse `json:"comment_child"`
	ReplyCommentID  *int              	`json:"reply_comment_id,omitempty"`
}


func (cs *CommentService) GetCommentsByPostID(postID string, userID string, orderStr string) ([]*CommentResponse, error) {
	var comments []models.Comment

	query := cs.db.
		Preload("User").
		Where("reply_post_id = ?", postID)

	// popular: 依 like_number；其他: 依 created_at (ASC/DESC)
	if orderStr == "popular" {
		query = query.Order("like_number DESC").Order("created_at DESC")
	} else {
		query = query.Order("created_at " + orderStr)
	}

	// 從資料庫撈出指定貼文的留言資料
	if err := query.Find(&comments).Error; err != nil {
		return nil, err
	}
	
	likedMap := make(map[int]bool)
	if userID != "" {
		var likedIDs []int
		if err := cs.db.Table("comment_likes cl").
			Select("cl.comment_id").
			Joins("JOIN comments c ON c.comment_id = cl.comment_id").
			Where("cl.user_id = ? AND c.reply_post_id = ?", userID, postID).
			Pluck("comment_id", &likedIDs).Error; err != nil {
			return nil, err
		}
		for _, id := range likedIDs {
			likedMap[id] = true
		}
	}

	
	
	commentMap := make(map[int]*CommentResponse)
	var roots []*CommentResponse

	for _, c := range comments {
		var pics []string
		if len(c.PictureURL) > 0 {
			if err := json.Unmarshal(c.PictureURL, &pics); err != nil {
				return nil, err
			}
		}

		commentMap[c.CommentID] = &CommentResponse{
			UserID:           c.UserID,
			UserFirstName:    c.User.FirstName,
			UserLastName:     c.User.LastName,
			ProfilePicture:	  c.User.ProfilePicture,
			CommentID:        c.CommentID,
			CommentContent:   c.Content,
			CommentTime:      c.CreatedAt.Format(time.RFC3339Nano),
			CommentPicture:   pics,
			CommentCount:     0,
			CommentLikeNumber: c.LikeNumber,
			ILikeItOrNot:     likedMap[c.CommentID],
			ReplyCommentID:   c.ReplyCommentID,
			CommentChild:     []CommentResponse{},
		}
	}
	// 在第一個 loop 結束後加這行
	
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
		//fmt.Println("Response sample:", *roots[0])
	}
	return roots, nil
}

func (cs *CommentService) CreateComment(postIDInt int, reply_comment_id string, userID string, content string, pic_files []*multipart.FileHeader,) (*CommentResponse, error) {
	
	//這邊要處理存照片到gcs並吐回網址
	pictureURLs, err := cs.uploadCommentPicturesToGCS(context.Background(), userID, pic_files)
	if err != nil {
		return nil, err
	}

	pictureJSON, err := json.Marshal(pictureURLs) //把GO的list轉成json二進位格式
	if err != nil {
		return nil, fmt.Errorf("failed to encode pictures: %w", err)
	}

	comment := models.Comment{
		UserID:      userID,
		ReplyPostID: &postIDInt,
		Content:     content,
		PictureURL:  datatypes.JSON(pictureJSON), //二進位再轉成文字版json（就是我們平常看到的）
		CreatedAt:   time.Now().UTC(),
	}
	
	if reply_comment_id != "" {
		var fatherID int
		fmt.Sscanf(reply_comment_id, "%d", &fatherID)
		comment.ReplyCommentID = &fatherID
	}

	//存進資料庫
	if err := cs.db.Create(&comment).Error; err != nil {
		return nil, err
	}

	return &CommentResponse{
		UserID:         comment.UserID,
		CommentID:      comment.CommentID,
		CommentContent: comment.Content,
		CommentPicture: pictureURLs,
		CommentTime: comment.CreatedAt.UTC().Format(time.RFC3339),
		ILikeItOrNot:   false,
		ReplyCommentID: comment.ReplyCommentID,

	}, nil
}

func (cs *CommentService) uploadCommentPicturesToGCS(ctx context.Context, userID string, files []*multipart.FileHeader) ([]string, error) {
	if len(files) == 0 {
		return []string{}, nil
	}
	if cs.gcsClient == nil {
		return nil, fmt.Errorf("gcs client is not initialized")
	}

	bucketName := os.Getenv("GCS_BUCKET_NAME")
	if bucketName == "" {
		return nil, fmt.Errorf("GCS_BUCKET_NAME is not configured")
	}

	userSegment := sanitizePathSegment(userID)
	pictureURLs := make([]string, 0, len(files))

	for _, file := range files {
		ext := filepath.Ext(file.Filename)
		objectName := fmt.Sprintf("comments/%s/%s%s", userSegment, uuid.NewString(), ext)

		src, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open upload file: %w", err)
		}

		writer := cs.gcsClient.Bucket(bucketName).Object(objectName).NewWriter(ctx)
		if contentType := file.Header.Get("Content-Type"); contentType != "" {
			writer.ContentType = contentType
		}

		if _, err := io.Copy(writer, src); err != nil {
			_ = src.Close()
			_ = writer.Close()
			return nil, fmt.Errorf("failed to upload image to GCS: %w", err)
		}

		if err := src.Close(); err != nil {
			_ = writer.Close()
			return nil, fmt.Errorf("failed to close upload file: %w", err)
		}
		if err := writer.Close(); err != nil {
			return nil, fmt.Errorf("failed to finalize GCS upload: %w", err)
		}

		pictureURLs = append(pictureURLs, fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, objectName))
	}

	return pictureURLs, nil
}

func (cs *CommentService) LikeCommentByID(commentID int,userID string) error {

	like := models.CommentLike{
		UserID: userID,
		CommentID: commentID,
	}

	if err := cs.db.Create(&like).Error; err != nil {
		return err
	}
	// 更新資料庫中的 LikeNumber 欄位
	if err := cs.db.Model(&models.Comment{}).Where("comment_id = ?", commentID).UpdateColumn("like_number", gorm.Expr("like_number + ?", 1)).Error; err != nil {
		return err
	}
	return nil
}

func (cs *CommentService) UnlikeCommentByID(commentID int, userID string) error {
	if err := cs.db.Where("user_id = ? AND comment_id = ?", userID, commentID).Delete(&models.CommentLike{}).Error; err != nil {
		return err
	}

	// 更新資料庫中的 LikeNumber 欄位
	if err := cs.db.Model(&models.Comment{}).Where("comment_id = ?", commentID).UpdateColumn("like_number", gorm.Expr("like_number - ?", 1)).Error; err != nil {
		return err
	}
	return nil
}

func (cs *CommentService) ReportCommentByID(commentID int, reportReason string) error {
	return nil 
}

func (cs *CommentService) DeleteCommentByID(commentID int) error {
	if err := cs.db.Where("comment_id = ?", commentID).Delete(&models.Comment{}).Error; err != nil {
		return err
	}
	return nil
}
