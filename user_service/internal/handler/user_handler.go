package handler

import (
	"context"
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/OshakbayAigerim/read_space/user_service/internal/domain"
	"github.com/OshakbayAigerim/read_space/user_service/internal/usecase"
	pb "github.com/OshakbayAigerim/read_space/user_service/proto"
)

type UserHandler struct {
	pb.UnimplementedUserServiceServer
	uc usecase.UserUseCase
	nc *nats.Conn
}

func NewUserHandler(u usecase.UserUseCase, nc *nats.Conn) *UserHandler {
	return &UserHandler{uc: u, nc: nc}
}

func (h *UserHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	if req == nil || req.User == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	user := &domain.User{
		Name:     req.User.Name,
		Email:    req.User.Email,
		Password: req.User.Password,
	}
	created, err := h.uc.CreateUser(ctx, user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create user: %v", err)
	}

	evt := struct {
		Id    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}{
		Id:    created.ID.Hex(),
		Name:  created.Name,
		Email: created.Email,
	}
	if data, err := json.Marshal(evt); err == nil {
		if err := h.nc.Publish("user.created", data); err != nil {
			log.Printf("⚠ NATS publish error (user.created): %v", err)
		}
	}

	return &pb.UserResponse{
		User: &pb.User{
			Id:       created.ID.Hex(),
			Name:     created.Name,
			Email:    created.Email,
			Password: created.Password,
		},
	}, nil
}

func (h *UserHandler) GetUser(ctx context.Context, req *pb.UserID) (*pb.UserResponse, error) {
	if req == nil || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "user ID is required")
	}
	user, err := h.uc.GetUserByID(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}
	return &pb.UserResponse{
		User: &pb.User{
			Id:       user.ID.Hex(),
			Name:     user.Name,
			Email:    user.Email,
			Password: user.Password,
		},
	}, nil
}

func (h *UserHandler) ListAllUsers(_ *pb.Empty, stream pb.UserService_ListAllUsersServer) error {
	users, err := h.uc.ListUsers(stream.Context())
	if err != nil {
		return status.Errorf(codes.Internal, "cannot list users: %v", err)
	}
	for _, u := range users {
		if err := stream.Send(&pb.User{
			Id:       u.ID.Hex(),
			Name:     u.Name,
			Email:    u.Email,
			Password: u.Password,
		}); err != nil {
			return err
		}
	}
	return nil
}
