package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/OshakbayAigerim/read_space/order_service/internal/domain"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

const (
	orderCacheTTL = 15 * time.Minute
)

var (
	cacheHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "order_cache_hits_total",
			Help: "Number of cache hits",
		},
		[]string{"cache_type"},
	)
	cacheMisses = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "order_cache_misses_total",
			Help: "Number of cache misses",
		},
		[]string{"cache_type"},
	)
	cacheOperations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "order_cache_operations_total",
			Help: "Total number of cache operations",
		},
		[]string{"operation", "cache_type", "status"},
	)
	cacheLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "order_cache_latency_seconds",
			Help:    "Latency of cache operations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "cache_type"},
	)
)

func init() {
	prometheus.MustRegister(cacheHits, cacheMisses, cacheOperations, cacheLatency)
}

type OrderCache interface {
	Get(ctx context.Context, id string) (*domain.Order, error)
	Set(ctx context.Context, order *domain.Order) error
	Delete(ctx context.Context, id string) error
	SetByUser(ctx context.Context, userID string, orders []*domain.Order) error
	GetByUser(ctx context.Context, userID string) ([]*domain.Order, error)
	DeleteByUser(ctx context.Context, userID string) error
}

type orderCache struct {
	client *redis.Client
}

func NewOrderCache(client *redis.Client) OrderCache {
	return &orderCache{client: client}
}

func (c *orderCache) Get(ctx context.Context, id string) (*domain.Order, error) {
	start := time.Now()
	defer func() {
		cacheLatency.WithLabelValues("get", "order").Observe(time.Since(start).Seconds())
	}()

	key := c.orderKey(id)
	log.Printf("[CACHE] Trying to get order from cache, key: %s", key)
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			cacheMisses.WithLabelValues("order").Inc()
			cacheOperations.WithLabelValues("get", "order", "miss").Inc()
			log.Printf("[CACHE MISS] Order not found in cache, key: %s", key)
			return nil, nil
		}
		cacheOperations.WithLabelValues("get", "order", "error").Inc()
		log.Printf("[CACHE ERROR] Failed to get order from cache, key: %s, error: %v", key, err)
		return nil, fmt.Errorf("redis get error: %w", err)
	}

	cacheHits.WithLabelValues("order").Inc()
	cacheOperations.WithLabelValues("get", "order", "hit").Inc()
	log.Printf("[CACHE HIT] Order found in cache, key: %s", key)

	var order domain.Order
	if err := json.Unmarshal([]byte(val), &order); err != nil {
		cacheOperations.WithLabelValues("get", "order", "error").Inc()
		log.Printf("[CACHE ERROR] Failed to unmarshal order from cache, key: %s, error: %v", key, err)
		return nil, fmt.Errorf("json unmarshal error: %w", err)
	}
	log.Printf("[CACHE SUCCESS] Successfully retrieved order from cache, key: %s, orderID: %s", key, order.ID.Hex())
	return &order, nil
}

func (c *orderCache) Set(ctx context.Context, order *domain.Order) error {
	start := time.Now()
	defer func() {
		cacheLatency.WithLabelValues("set", "order").Observe(time.Since(start).Seconds())
	}()

	key := c.orderKey(order.ID.Hex())
	log.Printf("[CACHE] Trying to set order in cache, key: %s, orderID: %s", key, order.ID.Hex())
	val, err := json.Marshal(order)
	if err != nil {
		cacheOperations.WithLabelValues("set", "order", "error").Inc()
		log.Printf("[CACHE ERROR] Failed to marshal order for cache, key: %s, error: %v", key, err)
		return fmt.Errorf("json marshal error: %w", err)
	}

	if err := c.client.Set(ctx, key, val, orderCacheTTL).Err(); err != nil {
		cacheOperations.WithLabelValues("set", "order", "error").Inc()
		log.Printf("[CACHE ERROR] Failed to set order in cache, key: %s, error: %v", key, err)
		return fmt.Errorf("redis set error: %w", err)
	}

	cacheOperations.WithLabelValues("set", "order", "success").Inc()
	log.Printf("[CACHE SUCCESS] Successfully set order in cache, key: %s, orderID: %s, TTL: %v",
		key, order.ID.Hex(), orderCacheTTL)
	return nil
}

func (c *orderCache) Delete(ctx context.Context, id string) error {
	start := time.Now()
	defer func() {
		cacheLatency.WithLabelValues("delete", "order").Observe(time.Since(start).Seconds())
	}()

	key := c.orderKey(id)
	log.Printf("[CACHE] Trying to delete order from cache, key: %s", key)
	if err := c.client.Del(ctx, key).Err(); err != nil {
		cacheOperations.WithLabelValues("delete", "order", "error").Inc()
		log.Printf("[CACHE ERROR] Failed to delete order from cache, key: %s, error: %v", key, err)
		return fmt.Errorf("redis del error: %w", err)
	}

	cacheOperations.WithLabelValues("delete", "order", "success").Inc()
	log.Printf("[CACHE SUCCESS] Successfully deleted order from cache, key: %s", key)
	return nil
}

func (c *orderCache) SetByUser(ctx context.Context, userID string, orders []*domain.Order) error {
	start := time.Now()
	defer func() {
		cacheLatency.WithLabelValues("set", "user_orders").Observe(time.Since(start).Seconds())
	}()

	key := c.userOrdersKey(userID)
	log.Printf("[CACHE] Trying to set user orders in cache, key: %s, userID: %s", key, userID)
	val, err := json.Marshal(orders)
	if err != nil {
		cacheOperations.WithLabelValues("set", "user_orders", "error").Inc()
		log.Printf("[CACHE ERROR] Failed to marshal user orders for cache, key: %s, error: %v", key, err)
		return fmt.Errorf("json marshal error: %w", err)
	}

	if err := c.client.Set(ctx, key, val, orderCacheTTL).Err(); err != nil {
		cacheOperations.WithLabelValues("set", "user_orders", "error").Inc()
		log.Printf("[CACHE ERROR] Failed to set user orders in cache, key: %s, error: %v", key, err)
		return fmt.Errorf("redis set error: %w", err)
	}

	cacheOperations.WithLabelValues("set", "user_orders", "success").Inc()
	log.Printf("[CACHE SUCCESS] Successfully set user orders in cache, key: %s, userID: %s, orders count: %d",
		key, userID, len(orders))
	return nil
}

func (c *orderCache) GetByUser(ctx context.Context, userID string) ([]*domain.Order, error) {
	start := time.Now()
	defer func() {
		cacheLatency.WithLabelValues("get", "user_orders").Observe(time.Since(start).Seconds())
	}()

	key := c.userOrdersKey(userID)
	log.Printf("[CACHE] Trying to get user orders from cache, key: %s, userID: %s", key, userID)
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			cacheMisses.WithLabelValues("user_orders").Inc()
			cacheOperations.WithLabelValues("get", "user_orders", "miss").Inc()
			log.Printf("[CACHE MISS] User orders not found in cache, key: %s", key)
			return nil, nil
		}
		cacheOperations.WithLabelValues("get", "user_orders", "error").Inc()
		log.Printf("[CACHE ERROR] Failed to get user orders from cache, key: %s, error: %v", key, err)
		return nil, fmt.Errorf("redis get error: %w", err)
	}

	cacheHits.WithLabelValues("user_orders").Inc()
	cacheOperations.WithLabelValues("get", "user_orders", "hit").Inc()
	log.Printf("[CACHE HIT] User orders found in cache, key: %s", key)

	var orders []*domain.Order
	if err := json.Unmarshal([]byte(val), &orders); err != nil {
		cacheOperations.WithLabelValues("get", "user_orders", "error").Inc()
		log.Printf("[CACHE ERROR] Failed to unmarshal user orders from cache, key: %s, error: %v", key, err)
		return nil, fmt.Errorf("json unmarshal error: %w", err)
	}
	log.Printf("[CACHE SUCCESS] Successfully retrieved user orders from cache, key: %s, userID: %s, orders count: %d",
		key, userID, len(orders))
	return orders, nil
}

func (c *orderCache) DeleteByUser(ctx context.Context, userID string) error {
	start := time.Now()
	defer func() {
		cacheLatency.WithLabelValues("delete", "user_orders").Observe(time.Since(start).Seconds())
	}()

	key := c.userOrdersKey(userID)
	log.Printf("[CACHE] Trying to delete user orders from cache, key: %s, userID: %s", key, userID)
	if err := c.client.Del(ctx, key).Err(); err != nil {
		cacheOperations.WithLabelValues("delete", "user_orders", "error").Inc()
		log.Printf("[CACHE ERROR] Failed to delete user orders from cache, key: %s, error: %v", key, err)
		return fmt.Errorf("redis del error: %w", err)
	}

	cacheOperations.WithLabelValues("delete", "user_orders", "success").Inc()
	log.Printf("[CACHE SUCCESS] Successfully deleted user orders from cache, key: %s", key)
	return nil
}

func (c *orderCache) orderKey(id string) string {
	return fmt.Sprintf("order:%s", id)
}

func (c *orderCache) userOrdersKey(userID string) string {
	return fmt.Sprintf("user_orders:%s", userID)
}
