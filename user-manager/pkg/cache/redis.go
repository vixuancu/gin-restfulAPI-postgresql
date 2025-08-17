package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCacheService struct {
	ctx context.Context
	rdb *redis.Client
}

func NewRedisCacheService(rdb *redis.Client) *RedisCacheService {
	return &RedisCacheService{
		ctx: context.Background(),
		rdb: rdb,
	}
}

func (cs *RedisCacheService) Set(key string, value any, expiration time.Duration) error {
	data, err := json.Marshal(value) // Chuyển đổi giá trị thành JSON
	if err != nil {
		return err
	}
	return cs.rdb.Set(cs.ctx, key, data, expiration).Err()

}
func (cs *RedisCacheService) Get(key string, dest any) error {
	data, err := cs.rdb.Get(cs.ctx, key).Result()
	if err == redis.Nil {
		return nil // Key không tồn tại
	}
	if err != nil {
		return err // Lỗi khác
	}
	return json.Unmarshal([]byte(data), dest) // Chuyển đổi JSON về giá trị gốc
}
func (cs *RedisCacheService) Clear(pattern string) error {
	cursor := uint64(0) // Khởi tạo con trỏ để quét
	for {
		keys, nextCursor, err := cs.rdb.Scan(cs.ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err // Lỗi quét
		}
		if len(keys) > 0 {
		if err := cs.rdb.Del(cs.ctx, keys...).Err(); err != nil {
        	return err
    }
		}
		if nextCursor == 0 {
			break // Kết thúc quét khi con trỏ trở về 0
		}
		cursor = nextCursor // Cập nhật con trỏ
	}
	return nil // Trả về nil nếu không có lỗi
}