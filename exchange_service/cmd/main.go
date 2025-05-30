package main

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
	"net"
	"time"

	"github.com/OshakbayAigerim/read_space/exchange_service/internal/cache"
	"github.com/OshakbayAigerim/read_space/exchange_service/internal/config"
	"github.com/OshakbayAigerim/read_space/exchange_service/internal/handler"
	"github.com/OshakbayAigerim/read_space/exchange_service/internal/repository"
	"github.com/OshakbayAigerim/read_space/exchange_service/internal/usecase"
	exchangepb "github.com/OshakbayAigerim/read_space/exchange_service/proto"
	userlibpb "github.com/OshakbayAigerim/read_space/user_library_service/proto"
	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
)

func main() {
	mongoClient := config.ConnectMongo()
	db := mongoClient.Database("readspace")

	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Redis connect error: %v", err)
	}
	log.Println(" Connected to Redis")

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf("NATS connect error: %v", err)
	}
	defer nc.Close()

	libConn, err := grpc.Dial("localhost:50055", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("cannot dial UserLibraryService: %v", err)
	}
	defer libConn.Close()
	libClient := userlibpb.NewUserLibraryServiceClient(libConn)

	repo := repository.NewMongoExchangeRepository(db)
	redisCache := cache.NewRedisExchangeCache(repo, rdb, 5*time.Minute)

	uc := usecase.NewExchangeUseCase(repo, redisCache, libClient)
	srv := handler.NewExchangeHandler(uc, nc)

	lis, err := net.Listen("tcp", ":50054")
	if err != nil {
		log.Fatalf("listen error: %v", err)
	}
	grpcServer := grpc.NewServer()
	exchangepb.RegisterExchangeServiceServer(grpcServer, srv)

	log.Println("ExchangeService listening on :50054")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve error: %v", err)
	}
}
