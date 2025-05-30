package handler

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/OshakbayAigerim/read_space/user_library_service/internal/domain"
	"github.com/OshakbayAigerim/read_space/user_library_service/internal/usecase"
	userpb "github.com/OshakbayAigerim/read_space/user_library_service/proto"
)

type UserLibraryHandler struct {
	userpb.UnimplementedUserLibraryServiceServer
	uc usecase.UserLibraryUseCase
	nc *nats.Conn
}

func NewUserLibraryHandler(uc usecase.UserLibraryUseCase, nc *nats.Conn) *UserLibraryHandler {
	return &UserLibraryHandler{uc: uc, nc: nc}
}

func mustMarshal(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}

func toProto(u *domain.UserBook) *userpb.UserBook {
	return &userpb.UserBook{
		Id:     u.ID.Hex(),
		UserId: u.UserID.Hex(),
		BookId: u.BookID.Hex(),
	}
}

func toProtoList(src []*domain.UserBook) []*userpb.UserBook {
	dst := make([]*userpb.UserBook, len(src))
	for i, u := range src {
		dst[i] = toProto(u)
	}
	return dst
}

func (h *UserLibraryHandler) AssignBook(ctx context.Context, req *userpb.AssignBookRequest) (*userpb.AssignBookResponse, error) {
	if req.UserId == "" || req.BookId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and book_id are required")
	}
	entry, err := h.uc.AssignBook(ctx, req.UserId, req.BookId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot assign book: %v", err)
	}
	evt := domain.BookAssignedEvent{UserID: req.UserId, BookID: req.BookId}
	h.nc.Publish("userlibrary.book.assigned", mustMarshal(evt))
	return &userpb.AssignBookResponse{Entry: toProto(entry)}, nil
}

func (h *UserLibraryHandler) UnassignBook(ctx context.Context, req *userpb.UnassignBookRequest) (*userpb.UnassignBookResponse, error) {
	if req.UserId == "" || req.BookId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and book_id are required")
	}
	if err := h.uc.UnassignBook(ctx, req.UserId, req.BookId); err != nil {
		return nil, status.Errorf(codes.Internal, "cannot unassign book: %v", err)
	}
	evt := domain.BookUnassignedEvent{UserID: req.UserId, BookID: req.BookId}
	h.nc.Publish("userlibrary.book.unassigned", mustMarshal(evt))
	return &userpb.UnassignBookResponse{Success: true}, nil
}

func (h *UserLibraryHandler) ListUserBooks(ctx context.Context, req *userpb.ListUserBooksRequest) (*userpb.ListUserBooksResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	list, err := h.uc.ListUserBooks(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot list user books: %v", err)
	}
	return &userpb.ListUserBooksResponse{Entries: toProtoList(list)}, nil
}

func (h *UserLibraryHandler) GetEntry(ctx context.Context, req *userpb.GetEntryRequest) (*userpb.AssignBookResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	e, err := h.uc.GetEntry(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "entry not found: %v", err)
	}
	return &userpb.AssignBookResponse{Entry: toProto(e)}, nil
}

func (h *UserLibraryHandler) DeleteEntry(ctx context.Context, req *userpb.DeleteEntryRequest) (*userpb.UnassignBookResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	// прежде чем удалить, получим user_id для события
	e, err := h.uc.GetEntry(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "entry not found: %v", err)
	}
	if err := h.uc.DeleteEntry(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "cannot delete entry: %v", err)
	}
	evt := domain.EntryDeletedEvent{EntryID: req.Id, UserID: e.UserID.Hex()}
	h.nc.Publish("userlibrary.entry.deleted", mustMarshal(evt))
	return &userpb.UnassignBookResponse{Success: true}, nil
}

func (h *UserLibraryHandler) UpdateEntry(ctx context.Context, req *userpb.UpdateEntryRequest) (*userpb.AssignBookResponse, error) {
	if req.Entry == nil || req.Entry.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "entry with id is required")
	}
	oid, err := primitive.ObjectIDFromHex(req.Entry.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid entry id")
	}
	uo, err := primitive.ObjectIDFromHex(req.Entry.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	bo, err := primitive.ObjectIDFromHex(req.Entry.BookId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid book_id")
	}
	dom := &domain.UserBook{ID: oid, UserID: uo, BookID: bo}
	updated, err := h.uc.UpdateEntry(ctx, dom)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot update entry: %v", err)
	}
	evt := domain.EntryUpdatedEvent{EntryID: req.Entry.Id, UserID: req.Entry.UserId, BookID: req.Entry.BookId}
	h.nc.Publish("userlibrary.entry.updated", mustMarshal(evt))
	return &userpb.AssignBookResponse{Entry: toProto(updated)}, nil
}

func (h *UserLibraryHandler) ListAllEntries(ctx context.Context, _ *emptypb.Empty) (*userpb.ListUserBooksResponse, error) {
	all, err := h.uc.ListAllEntries(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot list all entries: %v", err)
	}
	return &userpb.ListUserBooksResponse{Entries: toProtoList(all)}, nil
}

func (h *UserLibraryHandler) ListByBook(ctx context.Context, req *userpb.ListByBookRequest) (*userpb.ListUserBooksResponse, error) {
	if req.BookId == "" {
		return nil, status.Error(codes.InvalidArgument, "book_id is required")
	}
	list, err := h.uc.ListByBook(ctx, req.BookId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot list by book: %v", err)
	}
	return &userpb.ListUserBooksResponse{Entries: toProtoList(list)}, nil
}
