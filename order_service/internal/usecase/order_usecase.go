package usecase

import (
	"context"

	"github.com/OshakbayAigerim/read_space/order_service/internal/domain"
	"github.com/OshakbayAigerim/read_space/order_service/internal/repository"
)

type OrderUseCase interface {
	CreateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error)
	GetOrderByID(ctx context.Context, id string) (*domain.Order, error)
	ListOrdersByUser(ctx context.Context, userID string) ([]*domain.Order, error)
	CancelOrder(ctx context.Context, id string) (*domain.Order, error)
	ReturnBook(ctx context.Context, id string) (*domain.Order, error)
	DeleteOrder(ctx context.Context, id string) error
}

type orderUseCase struct {
	repo repository.OrderRepository
}

func NewOrderUseCase(r repository.OrderRepository) OrderUseCase {
	return &orderUseCase{repo: r}
}

func (u *orderUseCase) CreateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	return u.repo.Create(ctx, order)
}

func (u *orderUseCase) GetOrderByID(ctx context.Context, id string) (*domain.Order, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *orderUseCase) ListOrdersByUser(ctx context.Context, userID string) ([]*domain.Order, error) {
	return u.repo.ListByUser(ctx, userID)
}

func (u *orderUseCase) CancelOrder(ctx context.Context, id string) (*domain.Order, error) {
	return u.repo.Cancel(ctx, id)
}

func (u *orderUseCase) ReturnBook(ctx context.Context, id string) (*domain.Order, error) {
	return u.repo.Return(ctx, id)
}

func (u *orderUseCase) DeleteOrder(ctx context.Context, id string) error {
	return u.repo.Delete(ctx, id)
}
