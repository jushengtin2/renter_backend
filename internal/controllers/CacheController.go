package controllers

import (
	"net/http"
	"renter_backend/internal/services"

	"github.com/gin-gonic/gin"
)

type CacheController struct {
	cache_service *services.CacheService
}

func NewCacheController(s *services.CacheService) *CacheController {
	return &CacheController{cache_service: s}
}

func (cc *CacheController) WarmupCache(c *gin.Context) {
	userID:= c.GetString("clerk_user_id") // 從 context 取得 user id
	err := cc.cache_service.WarmupCache(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Cache warmup completed"})
}