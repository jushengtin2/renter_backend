package internal

import (
	"renter_backend/internal/controllers"
	"renter_backend/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
    r := gin.Default()

    // 建立 Service
    postService := services.NewPostService(db)

    // 建立 Controller，注入 Service
    postController := controllers.NewPostController(postService)

    // 路由綁定
    api := r.Group("/api/v1")
    {
        api.GET("/posts", postController.GetPostForMainPage)   // main page多筆
        api.GET("/posts/:postid", postController.GetPostByID)  // 單筆
    }

    return r
}
