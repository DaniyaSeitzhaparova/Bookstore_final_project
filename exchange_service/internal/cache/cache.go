package cache

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"time"

	"github.com/OshakbayAigerim/read_space/exchange_service/internal/domain"
	"github.com/OshakbayAigerim/read_space/exchange_service/internal/repository"
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
	if data, err := c.rdb.Get(ctx, key).Bytes(); err == nil {
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
	if blob, err := json.Marshal(offers); err == nil {
		_ = c.rdb.Set(ctx, key, blob, c.ttl).Err()
	}
	return offers, nil
}

func (c *RedisExchangeCache) ListPending(ctx context.Context) ([]*domain.ExchangeOffer, error) {
	key := "exchange:offers:pending"
	if data, err := c.rdb.Get(ctx, key).Bytes(); err == nil {
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
	if blob, err := json.Marshal(offers); err == nil {
		_ = c.rdb.Set(ctx, key, blob, c.ttl).Err()
	}
	return offers, nil
}

func (c *RedisExchangeCache) InvalidateUser(ctx context.Context, userID string) error {
	return c.rdb.Del(ctx, "exchange:offers:user:"+userID).Err()
}

func (c *RedisExchangeCache) InvalidatePending(ctx context.Context) error {
	return c.rdb.Del(ctx, "exchange:offers:pending").Err()
}
