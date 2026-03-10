package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB         *gorm.DB         // 全域變數 - 資料庫
	Storage_Client  *storage.Client  // 全域變數 - GCS Client
	BucketName string           // GCS Bucket 名稱
)

// Connect 初始化資料庫和 GCS 連線
func Connect() (*gorm.DB, *storage.Client, error) {
	log.Println("資料庫連線開始")

	appEnv := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	isCloud := appEnv == "cloud" || appEnv == "production"
	var dsn string
	// === 1. 初始化資料庫連線 ===
	// 雲端：強制使用 Supabase 連線字串
	// 本地：優先吃 SUPABASE_DB_URL；若沒設，才 fallback 到 DB_* 設定
	if isCloud{
		dsn = strings.TrimSpace(os.Getenv("SUPABASE_DB_URL"))
		println("prod mode")
	}

	if isCloud && dsn == "" {
		return nil, nil, fmt.Errorf("APP_ENV=%s but SUPABASE_DB_URL is empty", appEnv)
	}
	if dsn == "" {
		println("dev mode")
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		pass := os.Getenv("DB_PASSWORD")
		name := os.Getenv("DB_NAME")
		ssl := os.Getenv("DB_SSLMODE")
		tz := os.Getenv("DB_TIMEZONE")

		dsn = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
			host, user, pass, name, port, ssl, tz,
		)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect database: %v", err)
	}

	DB = db
	log.Println("資料庫連線成功")

	// === 2. 初始化 GCS Client ===
	// // 雲端模式：不使用 GCS，只使用 Supabase（DB）
	// if isCloud {
	// 	Storage_Client = nil
	// 	BucketName = strings.TrimSpace(os.Getenv("SUPABASE_STORAGE_BUCKET"))
	// 	if BucketName == "" {
	// 		BucketName = strings.TrimSpace(os.Getenv("BUCKET_NAME"))
	// 	}
	// 	log.Println("雲端模式：跳過 GCS 初始化，僅使用 Supabase")
	// 	return DB, nil, nil
	// }

	// 下面是storage (GCS) 初始化流程
	log.Println("ＧＣＳ連線開始")
	ctx := context.Background()
	
	gcsClient, err := storage.NewClient(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create GCS client: %v", err)
	}

	Storage_Client = gcsClient

	// 取得 Bucket 名稱
	BucketName = strings.TrimSpace(os.Getenv("GCS_BUCKET_NAME"))
	if BucketName == "" {
		log.Println("警告: GCS_BUCKET_NAME 環境變數未設定")
	} else {
		log.Printf("GCS BUCKET Client 初始化成功 (Bucket: %s)\n", BucketName)
	}

	return DB, Storage_Client, nil
}

// Close 關閉所有連線
func Close() {
	// 關閉 GCS Client
	if Storage_Client != nil {
		if err := Storage_Client.Close(); err != nil {
			log.Printf("關閉 GCS Client 失敗: %v", err)
		} else {
			log.Println("GCS Client 已關閉")
		}
	}

	// 關閉資料庫連線
	if DB != nil {
		sqlDB, err := DB.DB()
		if err == nil {
			if err := sqlDB.Close(); err != nil {
				log.Printf("關閉資料庫失敗: %v", err)
			} else {
				log.Println("資料庫連線已關閉")
			}
		}
	}
}
