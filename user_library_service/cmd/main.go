package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc"

	"github.com/OshakbayAigerim/read_space/user_library_service/internal/cache"
	"github.com/OshakbayAigerim/read_space/user_library_service/internal/config"
	"github.com/OshakbayAigerim/read_space/user_library_service/internal/handler"
	"github.com/OshakbayAigerim/read_space/user_library_service/internal/repository"
	"github.com/OshakbayAigerim/read_space/user_library_service/internal/usecase"
	userpb "github.com/OshakbayAigerim/read_space/user_library_service/proto"
)

func main() {
	// ——— Подключаемся к MongoDB ———
	mongoClient := config.ConnectMongo()
	db := mongoClient.Database("readspace")

	// —— DEBUG: сколько документов в коллекции сразу после подключения? ——
	count, err := db.Collection("user_books").CountDocuments(context.Background(), bson.M{})
	if err != nil {
		log.Fatalf("🔴 [DEBUG] cannot count user_books: %v", err)
	}
	log.Printf("🟢 [DEBUG] user_books collection has %d documents", count)
	// —————————————————————————————————————————————————————

	// ——— Подключаемся к Redis ———
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("🔴 Redis connect error: %v", err)
	}
	log.Println("🟢 Connected to Redis")

	// ——— Подключаемся к NATS ———
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf("🔴 NATS connect error: %v", err)
	}
	defer nc.Close()
	log.Println("🟢 Connected to NATS")

	// ——— Инициализируем слои ———
	repo := repository.NewMongoUserBookRepo(db)
	redisCache := cache.NewRedisUserLibraryCache(repo, rdb, 5*time.Minute)
	uc := usecase.NewUserLibraryUseCase(repo, redisCache)
	h := handler.NewUserLibraryHandler(uc, nc)

	// ——— Запускаем gRPC-сервер ———
	lis, err := net.Listen("tcp", ":50055")
	if err != nil {
		log.Fatalf("🔴 failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	userpb.RegisterUserLibraryServiceServer(grpcServer, h)

	log.Println("🟢 UserLibraryService listening on :50055")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("🔴 failed to serve: %v", err)
	}
}
