// internal/config/config.go
package config

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectMongo() *mongo.Client {
	uri := "mongodb://localhost:27017"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf(" MongoDB connection error: %v", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf(" MongoDB ping error: %v", err)
	}

	log.Println(" Connected to MongoDB")
	return client
}

func ConnectRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Redis connection error: %v", err)
	}

	log.Println("Connected to Redis")
	return client
}
