package repository

import (
	"context"
	"github.com/OshakbayAigerim/read_space/user_library_service/internal/domain"
)

type UserBookRepo interface {
	AssignBook(ctx context.Context, entry *domain.UserBook) (*domain.UserBook, error)
	UnassignBook(ctx context.Context, userID, bookID string) error
	ListUserBooks(ctx context.Context, userID string) ([]*domain.UserBook, error)
}
