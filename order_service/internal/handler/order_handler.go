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
		ID:      primitive.NewObjectID(),
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
	if raw, err := json.Marshal(evt); err == nil {
		if err := h.nc.Publish("orders.created", raw); err != nil {
			log.Printf("⚠️ publish orders.created: %v", err)
		}
	}

	return &pb.OrderResponse{Order: mapDomain(created)}, nil
}

func (h *OrderHandler) GetOrder(ctx context.Context, req *pb.OrderID) (*pb.OrderResponse, error) {
	if req == nil || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "order id is required")
	}
	ord, err := h.uc.GetOrderByID(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "order not found: %v", err)
	}
	return &pb.OrderResponse{Order: mapDomain(ord)}, nil
}

func (h *OrderHandler) ListOrdersByUser(ctx context.Context, req *pb.ListOrdersByUserRequest) (*pb.OrderList, error) {
	if req == nil || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	orders, err := h.uc.ListOrdersByUser(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot list orders: %v", err)
	}
	return &pb.OrderList{Orders: mapDomainList(orders)}, nil
}

func (h *OrderHandler) CancelOrder(ctx context.Context, req *pb.OrderID) (*pb.OrderResponse, error) {
	if req == nil || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "order id is required")
	}
	o, err := h.uc.CancelOrder(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot cancel order: %v", err)
	}
	return &pb.OrderResponse{Order: mapDomain(o)}, nil
}

func (h *OrderHandler) ReturnBook(ctx context.Context, req *pb.OrderID) (*pb.OrderResponse, error) {
	if req == nil || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "order id is required")
	}
	o, err := h.uc.ReturnBook(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot return order: %v", err)
	}
	return &pb.OrderResponse{Order: mapDomain(o)}, nil
}

func (h *OrderHandler) DeleteOrder(ctx context.Context, req *pb.OrderID) (*pb.Empty, error) {
	if req == nil || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "order id is required")
	}
	if err := h.uc.DeleteOrder(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "cannot delete order: %v", err)
	}
	return &pb.Empty{}, nil
}

func (h *OrderHandler) UpdateOrder(ctx context.Context, req *pb.UpdateOrderRequest) (*pb.OrderResponse, error) {
	if req == nil || req.Order == nil || req.Order.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "order object with id is required")
	}
	oid, err := primitive.ObjectIDFromHex(req.Order.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid order id")
	}
	uid, err := primitive.ObjectIDFromHex(req.Order.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	var bids []primitive.ObjectID
	for _, id := range req.Order.BookIds {
		bid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid book_id %q", id)
		}
		bids = append(bids, bid)
	}

	dom := &domain.Order{
		ID:      oid,
		UserID:  uid,
		BookIDs: bids,
		Status:  req.Order.Status,
	}
	updated, err := h.uc.UpdateOrder(ctx, dom)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot update order: %v", err)
	}
	return &pb.OrderResponse{Order: mapDomain(updated)}, nil
}

func (h *OrderHandler) AddBookToOrder(ctx context.Context, req *pb.BookOperationRequest) (*pb.OrderResponse, error) {
	if req == nil || req.OrderId == "" || req.BookId == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id and book_id are required")
	}
	o, err := h.uc.AddBook(ctx, req.OrderId, req.BookId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot add book to order: %v", err)
	}
	return &pb.OrderResponse{Order: mapDomain(o)}, nil
}

func (h *OrderHandler) RemoveBookFromOrder(ctx context.Context, req *pb.BookOperationRequest) (*pb.OrderResponse, error) {
	if req == nil || req.OrderId == "" || req.BookId == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id and book_id are required")
	}
	o, err := h.uc.RemoveBook(ctx, req.OrderId, req.BookId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot remove book from order: %v", err)
	}
	return &pb.OrderResponse{Order: mapDomain(o)}, nil
}

func (h *OrderHandler) ListAllOrders(ctx context.Context, _ *pb.Empty) (*pb.OrderList, error) {
	all, err := h.uc.ListAll(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot list all orders: %v", err)
	}
	return &pb.OrderList{Orders: mapDomainList(all)}, nil
}

func (h *OrderHandler) ListOrdersByStatus(ctx context.Context, req *pb.StatusRequest) (*pb.OrderList, error) {
	if req == nil || req.Status == "" {
		return nil, status.Error(codes.InvalidArgument, "status is required")
	}
	filtered, err := h.uc.ListByStatus(ctx, req.Status)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot list orders by status: %v", err)
	}
	return &pb.OrderList{Orders: mapDomainList(filtered)}, nil
}

func mapDomain(o *domain.Order) *pb.Order {
	var bookIDs []string
	for _, id := range o.BookIDs {
		bookIDs = append(bookIDs, id.Hex())
	}
	return &pb.Order{
		Id:        o.ID.Hex(),
		UserId:    o.UserID.Hex(),
		BookIds:   bookIDs,
		Status:    o.Status,
		CreatedAt: o.CreatedAt.Time().String(),
		UpdatedAt: o.UpdatedAt.Time().String(),
	}
}

func mapDomainList(list []*domain.Order) []*pb.Order {
	var out []*pb.Order
	for _, o := range list {
		out = append(out, mapDomain(o))
	}
	return out
}
