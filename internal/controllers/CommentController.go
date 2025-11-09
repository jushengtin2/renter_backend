package controllers

import (
	"fmt"
	"net/http"
	"renter_backend/internal/services"

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
	fmt.Println("UserID from context:", userID)
	userIDStr, ok := userID.(string)
	if !ok {
		// if the user id is missing or not a string, use empty string (or handle as needed)
		userIDStr = ""
	}
	comments, err := cc.CommentService.GetCommentsByPostID(postID, userIDStr)
	if err !=nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Println(comments)
	c.JSON(http.StatusOK, comments) 
}