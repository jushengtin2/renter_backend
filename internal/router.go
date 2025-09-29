package internal

import (
	"renter_backend/internal/controllers"
	"renter_backend/internal/services"
	"time"

	"renter_backend/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
    r := gin.Default()

    r.Use(cors.New(cors.Config{
    AllowAllOrigins:  true,           // 🔥 全部允許
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "ranking-page"},
    ExposeHeaders:    []string{"Content-Length"},
    AllowCredentials: false,          // 🚨 關掉，因為配合 *
    MaxAge:           12 * time.Hour,
}))


    // 建立 Service
    postService := services.NewPostService(db)

    // 建立 Controller，注入 Service
    postController := controllers.NewPostController(postService)

    // 路由綁定
    api := r.Group("/api/v1")
    {
        api.GET("/posts", postController.GetPostForMainPage)   // main page多筆
        api.GET("/posts/:postid", postController.GetPostByID)  // 單筆
        api.POST("/posts", middleware.ClerkAuth(), postController.CreatePost) // 發文需要驗證
        api.DELETE("/posts", middleware.ClerkAuth(), postController.DeletePost) // 刪文需要驗證
    }

  
    return r
}
