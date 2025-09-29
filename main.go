package main

import (
	"log"
	"renter_backend/internal"
	"renter_backend/internal/database"

	"github.com/gin-gonic/gin"
)

func main() {
    // 1. 連線資料庫
	db := database.Connect()

	// 2. 自動建立資料表
	database.AutoMigrate()

	// 3. 啟動 router，注入 db
    // 建立 Gin Router
    r := internal.SetupRouter(db)
    

    

    r.GET("/", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "BACKEND IS RUNNING",
        })
    })

    // 啟動伺服器
    if err := r.Run(":8080"); err != nil {
        log.Fatalf("server failed to start: %v", err)
    }
}
