package cache

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"time"

	"github.com/OshakbayAigerim/read_space/user_library_service/internal/domain"
	"github.com/OshakbayAigerim/read_space/user_library_service/internal/repository"
)

type UserLibraryCache interface {
	Get(ctx context.Context, userID string) ([]*domain.UserBook, error)
	Invalidate(ctx context.Context, userID string) error
}

type RedisUserLibraryCache struct {
	repo repository.UserBookRepo
	rdb  redis.UniversalClient
	ttl  time.Duration
}

func NewRedisUserLibraryCache(repo repository.UserBookRepo, rdb redis.UniversalClient, ttl time.Duration) *RedisUserLibraryCache {
	return &RedisUserLibraryCache{repo: repo, rdb: rdb, ttl: ttl}
}

func (c *RedisUserLibraryCache) Get(ctx context.Context, userID string) ([]*domain.UserBook, error) {
	key := "user_books:" + userID

	data, err := c.rdb.Get(ctx, key).Bytes()
	if err == nil {
		var entries []*domain.UserBook
		if err := json.Unmarshal(data, &entries); err == nil {
			return entries, nil
		}
	} else if err != redis.Nil {
		return nil, err
	}

	entries, err := c.repo.ListUserBooks(ctx, userID)
	if err != nil {
		return nil, err
	}

	if blob, err := json.Marshal(entries); err == nil {
		_ = c.rdb.Set(ctx, key, blob, c.ttl).Err()
	}

	return entries, nil
}

func (c *RedisUserLibraryCache) Invalidate(ctx context.Context, userID string) error {
	key := "user_books:" + userID
	return c.rdb.Del(ctx, key).Err()
}
