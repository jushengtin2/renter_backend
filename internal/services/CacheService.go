package services

import (
	"context"
	"fmt"
	"renter_backend/internal/models"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type CacheService struct {
	db *gorm.DB
	rdb *redis.Client
	ctx context.Context
}


func NewCacheService(db *gorm.DB, rdb *redis.Client) *CacheService {
	return &CacheService{
		db:  db,
		rdb: rdb,
		ctx: context.Background(),
	}
}


func (cs *CacheService) WarmupCache(userID string) error {
	var commentLikes []models.CommentLike

	// 從 DB 撈出該使用者按讚過的留言
	if err := cs.db.
		Where("user_id = ?", userID).
		Find(&commentLikes).Error; err != nil {
		return fmt.Errorf("撈取 comment_likes 失敗: %v", err)
	}

	// 組 Redis key
	key := userID

	// 先刪除舊的 key（避免重複或髒資料）
	if err := cs.rdb.Del(cs.ctx, userID).Err(); err != nil {
		return fmt.Errorf("刪除舊快取失敗: %v", err)
	}

	// 如果該使用者沒有按讚過任何留言，直接結束
	if len(commentLikes) == 0 {
		fmt.Printf("使用者 %s 沒有任何留言按讚紀錄，跳過快取。\n", userID)
		return nil
	}

	// 把所有 comment_id 加入 Redis Set
	members := make([]interface{}, len(commentLikes))
	for i, like := range commentLikes {
		
		members[i] = like.CommentID
	}

	if err := cs.rdb.SAdd(cs.ctx, key, members...).Err(); err != nil {
		return fmt.Errorf("寫入 Redis 失敗: %v", err)
	}

	// 設定 TTL (例如 3 小時)
	if err := cs.rdb.Expire(cs.ctx, key, 3*3600*1e9).Err(); err != nil { // 6hr
		return fmt.Errorf("設定 Redis TTL 失敗: %v", err)
	}

	fmt.Printf("✅ 使用者 %s 的留言按讚已快取，共 %d 筆。\n", userID, len(commentLikes))
	return nil
}
