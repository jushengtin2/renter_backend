package controllers

import (
	"fmt"
	"net/http"
	"renter_backend/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CommentController struct {
	services.CommentService
}

func NewCommentController (s *services.CommentService) *CommentController {
	return &CommentController{CommentService: *s}
}

func (cc *CommentController) GetCommentsByPostID(c *gin.Context) {
	postID := c.Param("postid")
	userID, _ := c.Get("clerk_user_id") 
	sort := c.Query("sort")
	var orderStr string

	userIDStr, ok := userID.(string)
	if !ok {
		// if the user id is missing or not a string, use empty string (or handle as needed)
		userIDStr = ""
	}

	if sort !="" && sort != "new" && sort != "popular" && sort!="old"{
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order"})
		return
	}
	if sort == "" || sort =="popular"{
		orderStr = "popular"
	}else if sort == "new"{
		orderStr = "DESC"
	} else{
		orderStr = "ASC"
	}


	comments, err := cc.CommentService.GetCommentsByPostID(postID, userIDStr, orderStr)
	if err !=nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Println(comments)
	c.JSON(http.StatusOK, comments) 
}

func (cc *CommentController) CreateComment(c *gin.Context) {
	postID := c.Param("postid")
	userID := c.GetString("clerk_user_id") //  // 已經從middleware context 取得 user id
	reply_comment_id := c.PostForm("reply_comment_id") // optional query parameter for nested comments
	comment_content := c.PostForm("content")
	
	postIDInt, err := strconv.Atoi(postID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid postid"})
		return
	}	

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart/form-data"})
		return
	}

	pic_files := form.File["picture_url"]
	if pic_files == nil && comment_content == ""{
		c.JSON(http.StatusBadRequest, gin.H{"error": "no data"})
		return
	}
	
	comment, err := cc.CommentService.CreateComment(postIDInt, reply_comment_id, userID, comment_content, pic_files)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, comment)
}

type CommentLikeRequest struct {
    CommentID string `json:"comment_id"` // 對應 JSON 中的 "postid" 鍵名
}

func (cc *CommentController) LikeCommentByID (c *gin.Context) {
	commentID := c.Param("commentid")
	userID := c.GetString("clerk_user_id") // 已經從middleware context 取得 user id
	
	commentIDInt, err := strconv.Atoi(commentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid postid"})
		return
	}
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	fmt.Println("userID from token:", userID)

	if err := cc.CommentService.LikeCommentByID(commentIDInt, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "post liked successfully"})

}

func (cc *CommentController) UnlikeCommentByID (c *gin.Context) {
	commentID := c.Param("commentid")
	userID := c.GetString("clerk_user_id") // 已經從middleware context 取得 user id
	
	commentIDInt, err := strconv.Atoi(commentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid postid"})
		return
	}
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	fmt.Println("userID from token:", userID)

	if err := cc.CommentService.UnlikeCommentByID(commentIDInt, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "post liked successfully"})

}

func (cc *CommentController) ReportCommentByID(c *gin.Context) {
	commentID := c.Param("commentid")
	reportReason := c.Query("reason") 
	if reportReason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "report reason is required"})
		return
	}
	commentID_int, err := strconv.Atoi(commentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid postid"})
		return
	}

	if err := cc.CommentService.ReportCommentByID(commentID_int, reportReason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "report successfully"})
}

func (cc *CommentController) DeleteCommentByID(c *gin.Context) {
	commentID := c.Param("commentid")
	if commentID == ""{
		c.JSON(http.StatusBadRequest, gin.H{"error": "commentid is required"})
	}
	commentIDint , err := strconv.Atoi(commentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid commentid"})
		return
	}
	
	if err := cc.CommentService.DeleteCommentByID(commentIDint); err != nil {
	    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	    return
	}
	c.JSON(http.StatusOK, gin.H{"message": "comment deleted successfully"})

}
