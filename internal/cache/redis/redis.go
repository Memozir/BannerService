package redis

import (
	"context"
	"fmt"
	"github.com/Memozir/BannerService/config"
	redis_ "github.com/redis/go-redis/v9"
	"time"
)

type RedisCache struct {
	redis *redis_.Client
}

func NewRedis(cfg *config.Config) *RedisCache {
	redis := redis_.NewClient(&redis_.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPass,
		DB:       cfg.RedisDb,
	})

	redisCache := RedisCache{redis: redis}

	return &redisCache
}

func (cache *RedisCache) GetBanner(ctx context.Context, featureId string, tagId string) (string, error) {
	key := fmt.Sprintf("%s:%s", featureId, tagId)
	res := cache.redis.Get(ctx, key)
	return res.String(), res.Err()
}

func (cache *RedisCache) SetBanner(ctx context.Context, featureId string, tagID string, banner string) error {
	key := fmt.Sprintf("%s:%s", featureId, tagID)
	res := cache.redis.Set(ctx, key, banner, 5*time.Minute)
	return res.Err()
}
