package handler

import (
	"context"
	"encoding/json"

	"github.com/OshakbayAigerim/read_space/user_library_service/internal/usecase"
	userpb "github.com/OshakbayAigerim/read_space/user_library_service/proto"
	"github.com/nats-io/nats.go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserLibraryHandler struct {
	userpb.UnimplementedUserLibraryServiceServer
	uc usecase.UserLibraryUseCase
	nc *nats.Conn
}

func NewUserLibraryHandler(uc usecase.UserLibraryUseCase, nc *nats.Conn) *UserLibraryHandler {
	return &UserLibraryHandler{uc: uc, nc: nc}
}

func (h *UserLibraryHandler) AssignBook(ctx context.Context, req *userpb.AssignBookRequest) (*userpb.AssignBookResponse, error) {
	if req.UserId == "" || req.BookId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and book_id are required")
	}
	entry, err := h.uc.AssignBook(ctx, req.UserId, req.BookId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot assign book: %v", err)
	}

	evt := struct {
		UserID string `json:"user_id"`
		BookID string `json:"book_id"`
	}{UserID: req.UserId, BookID: req.BookId}
	if data, _ := json.Marshal(evt); data != nil {
		h.nc.Publish("userlibrary.book.assigned", data)
	}

	return &userpb.AssignBookResponse{
		Entry: &userpb.UserBook{
			Id:     entry.ID.Hex(),
			UserId: entry.UserID.Hex(),
			BookId: entry.BookID.Hex(),
		},
	}, nil
}

func (h *UserLibraryHandler) UnassignBook(ctx context.Context, req *userpb.UnassignBookRequest) (*userpb.UnassignBookResponse, error) {
	if req.UserId == "" || req.BookId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and book_id are required")
	}
	if err := h.uc.UnassignBook(ctx, req.UserId, req.BookId); err != nil {
		return nil, status.Errorf(codes.Internal, "cannot unassign book: %v", err)
	}

	evt := struct {
		UserID string `json:"user_id"`
		BookID string `json:"book_id"`
	}{UserID: req.UserId, BookID: req.BookId}
	if data, _ := json.Marshal(evt); data != nil {
		h.nc.Publish("userlibrary.book.unassigned", data)
	}

	return &userpb.UnassignBookResponse{Success: true}, nil
}

func (h *UserLibraryHandler) ListUserBooks(ctx context.Context, req *userpb.ListUserBooksRequest) (*userpb.ListUserBooksResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	entries, err := h.uc.ListUserBooks(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot list user books: %v", err)
	}
	resp := &userpb.ListUserBooksResponse{}
	for _, e := range entries {
		resp.Entries = append(resp.Entries, &userpb.UserBook{
			Id:     e.ID.Hex(),
			UserId: e.UserID.Hex(),
			BookId: e.BookID.Hex(),
		})
	}
	return resp, nil
}
