package main

import (
	"context"
	"log"
	"net"

	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"

	"github.com/OshakbayAigerim/read_space/book_service/internal/cache"
	"github.com/OshakbayAigerim/read_space/book_service/internal/config"
	"github.com/OshakbayAigerim/read_space/book_service/internal/handler"
	"github.com/OshakbayAigerim/read_space/book_service/internal/repository"
	"github.com/OshakbayAigerim/read_space/book_service/internal/usecase"
	pb "github.com/OshakbayAigerim/read_space/book_service/proto"
)

func main() {
	ctx := context.Background()

	mongoClient := config.ConnectMongo()
	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting MongoDB: %v", err)
		}
	}()

	redisClient := config.ConnectRedis()
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing Redis connection: %v", err)
		}
	}()

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf(" NATS connect error: %v", err)
	}
	defer nc.Close()

	bookCache := cache.NewRedisBookCache(redisClient)

	bookRepo := repository.NewMongoBookRepository(mongoClient)
	cachedBookRepo := repository.NewCachedBookRepository(bookRepo, bookCache)

	bookUC := usecase.NewBookUseCase(cachedBookRepo)

	srv := handler.NewBookHandler(bookUC, nc)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf(" Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterBookServiceServer(grpcServer, srv)

	log.Println("BookService gRPC server started on port 50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf(" Failed to serve: %v", err)
	}
}
