package repository

import (
	"context"

	"github.com/OshakbayAigerim/read_space/user_library_service/internal/domain"
)

type UserBookRepo interface {
	AssignBook(ctx context.Context, entry *domain.UserBook) (*domain.UserBook, error)
	UnassignBook(ctx context.Context, userID, bookID string) error
	ListUserBooks(ctx context.Context, userID string) ([]*domain.UserBook, error)

	GetEntry(ctx context.Context, id string) (*domain.UserBook, error)
	DeleteEntry(ctx context.Context, id string) error
	UpdateEntry(ctx context.Context, entry *domain.UserBook) (*domain.UserBook, error)
	ListAllEntries(ctx context.Context) ([]*domain.UserBook, error)
	ListByBook(ctx context.Context, bookID string) ([]*domain.UserBook, error)
}
