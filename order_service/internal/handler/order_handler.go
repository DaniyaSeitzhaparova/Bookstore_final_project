package handler

import (
	"context"
	"encoding/json"
	"log"

	"github.com/OshakbayAigerim/read_space/order_service/internal/domain"
	"github.com/OshakbayAigerim/read_space/order_service/internal/usecase"
	pb "github.com/OshakbayAigerim/read_space/order_service/proto"
	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderHandler struct {
	pb.UnimplementedOrderServiceServer
	uc usecase.OrderUseCase
	nc *nats.Conn
}

func NewOrderHandler(u usecase.OrderUseCase, nc *nats.Conn) *OrderHandler {
	return &OrderHandler{uc: u, nc: nc}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.OrderResponse, error) {
	if req == nil || req.UserId == "" || len(req.BookIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id and book_ids are required")
	}

	uid, err := primitive.ObjectIDFromHex(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	var bids []primitive.ObjectID
	for _, id := range req.BookIds {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid book_id %q", id)
		}
		bids = append(bids, oid)
	}

	ord := &domain.Order{
		UserID:  uid,
		BookIDs: bids,
		Status:  "Created",
	}
	created, err := h.uc.CreateOrder(ctx, ord)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create order: %v", err)
	}

	evt := struct {
		OrderID string   `json:"order_id"`
		UserID  string   `json:"user_id"`
		BookIDs []string `json:"book_ids"`
	}{
		OrderID: created.ID.Hex(),
		UserID:  created.UserID.Hex(),
		BookIDs: req.BookIds,
	}
	raw, _ := json.Marshal(evt)
	if err := h.nc.Publish("orders.created", raw); err != nil {
		log.Printf("⚠Failed to publish orders.created: %v", err)
	}

	var respBookIDs []string
	for _, oid := range created.BookIDs {
		respBookIDs = append(respBookIDs, oid.Hex())
	}
	return &pb.OrderResponse{
		Order: &pb.Order{
			Id:        created.ID.Hex(),
			UserId:    created.UserID.Hex(),
			BookIds:   respBookIDs,
			Status:    created.Status,
			CreatedAt: created.CreatedAt.Time().String(),
			UpdatedAt: created.UpdatedAt.Time().String(),
		},
	}, nil
}

func (h *OrderHandler) ReturnBook(ctx context.Context, req *pb.OrderID) (*pb.OrderResponse, error) {
	if req == nil || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "order id is required")
	}
	o, err := h.uc.ReturnBook(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot return order: %v", err)
	}

	evt := struct {
		OrderID string   `json:"order_id"`
		UserID  string   `json:"user_id"`
		BookIDs []string `json:"book_ids"`
	}{
		OrderID: o.ID.Hex(),
		UserID:  o.UserID.Hex(),
		BookIDs: func(ids []primitive.ObjectID) []string {
			out := make([]string, len(ids))
			for i, x := range ids {
				out[i] = x.Hex()
			}
			return out
		}(o.BookIDs),
	}
	raw, _ := json.Marshal(evt)
	if err := h.nc.Publish("order.completed", raw); err != nil {
		log.Printf("⚠ Failed to publish order.completed: %v", err)
	}

	var bidStr []string
	for _, x := range o.BookIDs {
		bidStr = append(bidStr, x.Hex())
	}
	return &pb.OrderResponse{
		Order: &pb.Order{
			Id:        o.ID.Hex(),
			UserId:    o.UserID.Hex(),
			BookIds:   bidStr,
			Status:    o.Status,
			CreatedAt: o.CreatedAt.Time().String(),
			UpdatedAt: o.UpdatedAt.Time().String(),
		},
	}, nil
}

func (h *OrderHandler) DeleteOrder(ctx context.Context, req *pb.OrderID) (*pb.Empty, error) {
	if req == nil || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "order id is required")
	}

	o, err := h.uc.GetOrderByID(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "order not found: %v", err)
	}

	evt := struct {
		OrderID string   `json:"order_id"`
		UserID  string   `json:"user_id"`
		BookIDs []string `json:"book_ids"`
	}{
		OrderID: o.ID.Hex(),
		UserID:  o.UserID.Hex(),
		BookIDs: func(ids []primitive.ObjectID) []string {
			out := make([]string, len(ids))
			for i, x := range ids {
				out[i] = x.Hex()
			}
			return out
		}(o.BookIDs),
	}
	raw, _ := json.Marshal(evt)
	if err := h.nc.Publish("order.deleted", raw); err != nil {
		log.Printf("⚠ Failed to publish order.deleted: %v", err)
	}

	if err := h.uc.DeleteOrder(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "cannot delete order: %v", err)
	}
	return &pb.Empty{}, nil
}

func (h *OrderHandler) GetOrder(ctx context.Context, req *pb.OrderID) (*pb.OrderResponse, error) {
	if req == nil || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "order id is required")
	}
	ord, err := h.uc.GetOrderByID(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "order not found: %v", err)
	}
	var pbBids []string
	for _, bid := range ord.BookIDs {
		pbBids = append(pbBids, bid.Hex())
	}
	return &pb.OrderResponse{
		Order: &pb.Order{
			Id:        ord.ID.Hex(),
			UserId:    ord.UserID.Hex(),
			BookIds:   pbBids,
			Status:    ord.Status,
			CreatedAt: ord.CreatedAt.Time().String(),
			UpdatedAt: ord.UpdatedAt.Time().String(),
		},
	}, nil
}

func (h *OrderHandler) ListOrdersByUser(ctx context.Context, req *pb.ListOrdersByUserRequest) (*pb.OrderList, error) {
	if req == nil || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	orders, err := h.uc.ListOrdersByUser(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot list orders: %v", err)
	}
	var list []*pb.Order
	for _, o := range orders {
		var pbBids []string
		for _, bid := range o.BookIDs {
			pbBids = append(pbBids, bid.Hex())
		}
		list = append(list, &pb.Order{
			Id:        o.ID.Hex(),
			UserId:    o.UserID.Hex(),
			BookIds:   pbBids,
			Status:    o.Status,
			CreatedAt: o.CreatedAt.Time().String(),
			UpdatedAt: o.UpdatedAt.Time().String(),
		})
	}
	return &pb.OrderList{Orders: list}, nil
}

func (h *OrderHandler) CancelOrder(ctx context.Context, req *pb.OrderID) (*pb.OrderResponse, error) {
	if req == nil || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "order id is required")
	}
	o, err := h.uc.CancelOrder(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot cancel order: %v", err)
	}
	var bids []string
	for _, bid := range o.BookIDs {
		bids = append(bids, bid.Hex())
	}
	return &pb.OrderResponse{
		Order: &pb.Order{
			Id:        o.ID.Hex(),
			UserId:    o.UserID.Hex(),
			BookIds:   bids,
			Status:    o.Status,
			CreatedAt: o.CreatedAt.Time().String(),
			UpdatedAt: o.UpdatedAt.Time().String(),
		},
	}, nil
}
