package repository

import (
	"context"

	"github.com/OshakbayAigerim/read_space/order_service/internal/domain"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) (*domain.Order, error)
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	ListByUser(ctx context.Context, userID string) ([]*domain.Order, error)
	Cancel(ctx context.Context, id string) (*domain.Order, error)
	Return(ctx context.Context, id string) (*domain.Order, error)
	Update(ctx context.Context, order *domain.Order) (*domain.Order, error)
	AddBook(ctx context.Context, orderID, bookID string) (*domain.Order, error)
	RemoveBook(ctx context.Context, orderID, bookID string) (*domain.Order, error)
	ListAll(ctx context.Context) ([]*domain.Order, error)
	ListByStatus(ctx context.Context, status string) ([]*domain.Order, error)
	Delete(ctx context.Context, id string) error
}
