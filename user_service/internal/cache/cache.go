package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/OshakbayAigerim/read_space/user_service/internal/domain"
	"github.com/redis/go-redis/v9"
)

const (
	userCacheTTL = 15 * time.Minute
)

type UserCache struct {
	client *redis.Client
}

func NewUserCache(client *redis.Client) *UserCache {
	return &UserCache{client: client}
}

func (c *UserCache) Get(ctx context.Context, id string) (*domain.User, error) {
	data, err := c.client.Get(ctx, userKey(id)).Bytes()
	if err != nil {
		if err == redis.Nil {
			fmt.Printf("Cache MISS for user ID: %s\n", id)
			return nil, nil
		}
		return nil, err
	}

	var user domain.User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, err
	}

	fmt.Printf("Cache HIT for user ID: %s\n", id)
	return &user, nil
}

func (c *UserCache) Set(ctx context.Context, user *domain.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, userKey(user.ID.Hex()), data, userCacheTTL).Err()
}

func (c *UserCache) Delete(ctx context.Context, id string) error {
	return c.client.Del(ctx, userKey(id)).Err()
}

func userKey(id string) string {
	return "user:" + id
}
