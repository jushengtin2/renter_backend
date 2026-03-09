package internal

import (
	"renter_backend/internal/controllers"
	"renter_backend/internal/middleware"
	"renter_backend/internal/services"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB, rdb *redis.Client, gcsClient *storage.Client) *gin.Engine {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true, // 全部允許
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "ranking-page"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false, // 關掉，因為配合 *
		MaxAge:           12 * time.Hour,
	}))

	// 建立 Service
	postService := services.NewPostService(db, rdb, gcsClient)

	// 建立 Controller，注入 Service
	postController := controllers.NewPostController(postService)

	commentService := services.NewCommentService(db, rdb, gcsClient)
	commentController := controllers.NewCommentController(commentService)

	userService := services.NewUserService(db)
	userController := controllers.NewUserController(userService)
	// cacheService := services.NewCacheService(db, rdb)
	// cacheController := controllers.NewCacheController(cacheService)

	// 路由綁定
	api := r.Group("/api/v1")
	{	
		api.POST("/user", middleware.OptionalClerkAuth(), userController.GetUserProfile) // 這個路由是給clerk webhook用的，clerk會在user info更新的時候call這個路由，然後把新的user info傳過來，讓我們可以更新資料庫裡的user info
		api.GET("/posts", middleware.OptionalClerkAuth(), postController.GetPostForMainPage)  
		api.GET("/posts/:postid", middleware.OptionalClerkAuth(), postController.GetPostByID) 
		api.POST("/posts", middleware.ClerkAuth(), postController.CreatePost)             
		api.DELETE("/posts/:postid", middleware.ClerkAuth(), postController.DeletePost)             

		api.GET("/posts/:postid/comments", middleware.OptionalClerkAuth(), commentController.GetCommentsByPostID) 
		api.POST("/posts/:postid/comments", middleware.ClerkAuth(), commentController.CreateComment)    
		api.POST("/posts/:postid/comments/:commentid/like", middleware.ClerkAuth(), commentController.LikeCommentByID)      
		api.DELETE("/posts/:postid/comments/:commentid/like", middleware.ClerkAuth(), commentController.UnlikeCommentByID)    

		api.POST("/posts/:postid/like", middleware.ClerkAuth(), postController.LikePostByID)     
		api.DELETE("/posts/:postid/like", middleware.ClerkAuth(), postController.UnlikePostByID)  
		api.POST("/posts/:postid/save", middleware.ClerkAuth(), postController.SavePostByID)      
		api.DELETE("/posts/:postid/save", middleware.ClerkAuth(), postController.UnSavePostByID)    

		api.POST("/posts/:postid/report", middleware.OptionalClerkAuth(), postController.ReportPostByID) 

		api.POST("/comments/:commentid/report", middleware.OptionalClerkAuth(), commentController.ReportCommentByID) 
		api.DELETE("/comments/:commentid", middleware.ClerkAuth(), commentController.DeleteCommentByID)  

		api.GET("/posts/save", middleware.ClerkAuth(), postController.GetSavePost) 

		// api.POST("/cache/warmup", middleware.ClerkAuth(), cacheController.WarmupCache)
	}

	return r
}
