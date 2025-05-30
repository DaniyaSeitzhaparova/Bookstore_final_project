package main

import (
	"log"
	"net"
	"net/http"

	"github.com/OshakbayAigerim/read_space/order_service/internal/cache"
	"github.com/OshakbayAigerim/read_space/order_service/internal/config"
	"github.com/OshakbayAigerim/read_space/order_service/internal/handler"
	"github.com/OshakbayAigerim/read_space/order_service/internal/repository"
	"github.com/OshakbayAigerim/read_space/order_service/internal/usecase"
	pb "github.com/OshakbayAigerim/read_space/order_service/proto"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func main() {
	client := config.ConnectMongo()
	db := client.Database("readspace")

	redisClient := config.ConnectRedis()
	defer redisClient.Close()

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf("NATS connect error: %v", err)
	}
	defer nc.Close()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Println(" Metrics server started on :9091")
		log.Fatal(http.ListenAndServe(":9091", nil))
	}()

	orderCache := cache.NewOrderCache(redisClient)
	orderRepo := repository.NewMongoOrderRepository(db, orderCache)
	orderUC := usecase.NewOrderUseCase(orderRepo)

	h := handler.NewOrderHandler(orderUC, nc)

	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf(" failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterOrderServiceServer(grpcServer, h)

	log.Println("OrderService started on :50053")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
