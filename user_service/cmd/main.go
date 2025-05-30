package main

import (
	"github.com/OshakbayAigerim/read_space/user_service/internal/migration"
	"log"
	"net"

	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"

	"github.com/OshakbayAigerim/read_space/user_service/internal/cache"
	"github.com/OshakbayAigerim/read_space/user_service/internal/config"
	"github.com/OshakbayAigerim/read_space/user_service/internal/handler"
	"github.com/OshakbayAigerim/read_space/user_service/internal/repository"
	"github.com/OshakbayAigerim/read_space/user_service/internal/usecase"
	pb "github.com/OshakbayAigerim/read_space/user_service/proto"
)

func main() {
	client := config.ConnectMongo()
	db := client.Database("readspace")

	migrations.CreateUserCollectionIndexes(db)

	redisClient := config.ConnectRedis()
	userCache := cache.NewUserCache(redisClient)

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf(" NATS connect error: %v", err)
	}
	defer nc.Close()

	userRepo := repository.NewMongoUserRepository(db, userCache)
	userUC := usecase.NewUserUseCase(userRepo)
	srv := handler.NewUserHandler(userUC, nc)

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, srv)

	log.Println("UserService gRPC server started on port 50052")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
