package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/OshakbayAigerim/read_space/exchange_service/internal/domain"
	"github.com/OshakbayAigerim/read_space/exchange_service/internal/repository"
	"github.com/redis/go-redis/v9"
)

type ExchangeCache interface {
	ListByUser(ctx context.Context, userID string) ([]*domain.ExchangeOffer, error)
	ListPending(ctx context.Context) ([]*domain.ExchangeOffer, error)
	InvalidateUser(ctx context.Context, userID string) error
	InvalidatePending(ctx context.Context) error
}

type RedisExchangeCache struct {
	repo repository.ExchangeRepository
	rdb  redis.UniversalClient
	ttl  time.Duration
}

func NewRedisExchangeCache(repo repository.ExchangeRepository, rdb redis.UniversalClient, ttl time.Duration) *RedisExchangeCache {
	return &RedisExchangeCache{repo: repo, rdb: rdb, ttl: ttl}
}

func (c *RedisExchangeCache) ListByUser(ctx context.Context, userID string) ([]*domain.ExchangeOffer, error) {
	key := "exchange:offers:user:" + userID
	data, err := c.rdb.Get(ctx, key).Bytes()
	if err == nil {
		var offers []*domain.ExchangeOffer
		if json.Unmarshal(data, &offers) == nil {
			return offers, nil
		}
	} else if err != redis.Nil {
		return nil, err
	}

	offers, err := c.repo.ListOffersByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	blob, err := json.Marshal(offers)
	if err == nil {
		c.rdb.Set(ctx, key, blob, c.ttl)
	}
	return offers, nil
}

func (c *RedisExchangeCache) ListPending(ctx context.Context) ([]*domain.ExchangeOffer, error) {
	key := "exchange:offers:pending"
	data, err := c.rdb.Get(ctx, key).Bytes()
	if err == nil {
		var offers []*domain.ExchangeOffer
		if json.Unmarshal(data, &offers) == nil {
			return offers, nil
		}
	} else if err != redis.Nil {
		return nil, err
	}

	offers, err := c.repo.ListPendingOffers(ctx)
	if err != nil {
		return nil, err
	}
	blob, err := json.Marshal(offers)
	if err == nil {
		c.rdb.Set(ctx, key, blob, c.ttl)
	}
	return offers, nil
}

func (c *RedisExchangeCache) InvalidateUser(ctx context.Context, userID string) error {
	return c.rdb.Del(ctx, "exchange:offers:user:"+userID).Err()
}

func (c *RedisExchangeCache) InvalidatePending(ctx context.Context) error {
	return c.rdb.Del(ctx, "exchange:offers:pending").Err()
}
