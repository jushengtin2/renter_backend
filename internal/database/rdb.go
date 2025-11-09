package database

import (
	"fmt"

	"github.com/go-redis/redis/v8" // 我用redis來先存一些使用者資訊這樣以後可以減少去fetch db的次數（語法也比較簡潔）
)

func ConnectRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:8787", // Redis 伺服器地址（這個到時候要換因為我目前是把redis架在docker 8787:6379）
		Password: "",               // 沒有密碼設置
		DB:       0,                // 使用預設 DB
	})

	// 測試連接
	pong, err := rdb.Ping(rdb.Context()).Result()
	if err != nil {
		fmt.Println("無法連接到 Redis:", err)
	} else {
		fmt.Println("Redis 連接成功:", pong)
	}

	return rdb
}

