package main

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
	"net"
	"time"

	"github.com/OshakbayAigerim/read_space/user_library_service/internal/cache"
	"github.com/OshakbayAigerim/read_space/user_library_service/internal/config"
	"github.com/OshakbayAigerim/read_space/user_library_service/internal/handler"
	"github.com/OshakbayAigerim/read_space/user_library_service/internal/repository"
	"github.com/OshakbayAigerim/read_space/user_library_service/internal/usecase"
	userpb "github.com/OshakbayAigerim/read_space/user_library_service/proto"
	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
)

func main() {
	mongoClient := config.ConnectMongo()
	db := mongoClient.Database("readspace")

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf(" Redis connect error: %v", err)
	}
	log.Println(" Connected to Redis")

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf("NATS connect error: %v", err)
	}
	defer nc.Close()

	repo := repository.NewMongoUserBookRepo(db)
	redisCache := cache.NewRedisUserLibraryCache(repo, rdb, 5*time.Minute)

	uc := usecase.NewUserLibraryUseCase(repo, redisCache)
	h := handler.NewUserLibraryHandler(uc, nc)

	lis, err := net.Listen("tcp", ":50055")
	if err != nil {
		log.Fatalf(" failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	userpb.RegisterUserLibraryServiceServer(grpcServer, h)

	log.Println("UserLibraryService listening on :50055")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
