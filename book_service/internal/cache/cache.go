package cache

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"log"
	"time"

	"github.com/OshakbayAigerim/read_space/book_service/internal/domain"
)

type BookCache interface {
	Get(ctx context.Context, key string) (*domain.Book, error)
	Set(ctx context.Context, key string, book *domain.Book, expiration time.Duration) error

	GetList(ctx context.Context, key string) ([]*domain.Book, error)
	SetList(ctx context.Context, key string, books []*domain.Book, expiration time.Duration) error

	Delete(ctx context.Context, key string) error
}
type redisBookCache struct {
	client *redis.Client
}

func NewRedisBookCache(client *redis.Client) BookCache {
	return &redisBookCache{client: client}
}

func (r *redisBookCache) Get(ctx context.Context, key string) (*domain.Book, error) {
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var book domain.Book
	if err := json.Unmarshal([]byte(data), &book); err != nil {
		return nil, err
	}
	return &book, nil
}

func (r *redisBookCache) Set(ctx context.Context, key string, book *domain.Book, expiration time.Duration) error {
	data, err := json.Marshal(book)
	if err != nil {
		return err
	}

	err = r.client.Set(ctx, key, data, expiration).Err()
	if err != nil {
		return err
	}

	log.Printf("//Book with key '%s' cached successfully", key)
	return nil
}

func (r *redisBookCache) GetList(ctx context.Context, key string) ([]*domain.Book, error) {
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var books []*domain.Book
	if err := json.Unmarshal([]byte(data), &books); err != nil {
		return nil, err
	}
	return books, nil
}

func (r *redisBookCache) SetList(ctx context.Context, key string, books []*domain.Book, expiration time.Duration) error {
	data, err := json.Marshal(books)
	if err != nil {
		return err
	}

	err = r.client.Set(ctx, key, data, expiration).Err()
	if err != nil {
		return err
	}

	log.Printf("Book list with key '%s' cached successfully", key)
	return nil
}

func (r *redisBookCache) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return err
	}

	log.Printf("🗑 Cache key '%s' deleted", key)
	return nil
}
