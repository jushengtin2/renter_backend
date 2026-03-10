package main

import (
	"log"
	"os"
	"renter_backend/internal"
	"renter_backend/internal/database"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

    //0. 載入 .env
    _ = godotenv.Load()

	// 1. 連線資料庫
	db, supabaseClient, err := database.Connect()
	if err != nil {
		log.Fatalf("database connect failed: %v", err)
	}

	// 2. 自動建立資料表
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}

	// 3. 啟動 router，注入 db
    // 建立 Gin Router
    r := internal.SetupRouter(db, supabaseClient)
    
    r.GET("/", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "BACKEND IS RUNNING",
        })
    })

    // 啟動伺服器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
	
}
