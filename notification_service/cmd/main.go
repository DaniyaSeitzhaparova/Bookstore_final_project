package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"

	"github.com/OshakbayAigerim/read_space/notification_service/internal/config"
	"github.com/OshakbayAigerim/read_space/notification_service/internal/handler"
	"github.com/OshakbayAigerim/read_space/notification_service/internal/usecase"
	userpb "github.com/OshakbayAigerim/read_space/user_service/proto"
)

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf("NATS connect error: %v", err)
	}
	defer nc.Close()

	conn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial UserService: %v", err)
	}
	defer conn.Close()
	userClient := userpb.NewUserServiceClient(conn)

	notifier := usecase.NewNotifier(userClient, config.SendEmail)

	if err := handler.SubscribeAll(nc, notifier); err != nil {
		log.Fatalf(" failed to subscribe: %v", err)
	}
	log.Println("NotificationService subscribed to all relevant events")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	log.Println("NotificationService shutting down")
}
