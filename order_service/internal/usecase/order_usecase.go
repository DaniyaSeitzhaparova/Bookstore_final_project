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
	UpdateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error)
	AddBook(ctx context.Context, orderID, bookID string) (*domain.Order, error)
	RemoveBook(ctx context.Context, orderID, bookID string) (*domain.Order, error)
	ListAll(ctx context.Context) ([]*domain.Order, error)
	ListByStatus(ctx context.Context, status string) ([]*domain.Order, error)
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

func (u *orderUseCase) UpdateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	return u.repo.Update(ctx, order)
}

func (u *orderUseCase) AddBook(ctx context.Context, orderID, bookID string) (*domain.Order, error) {
	return u.repo.AddBook(ctx, orderID, bookID)
}

func (u *orderUseCase) RemoveBook(ctx context.Context, orderID, bookID string) (*domain.Order, error) {
	return u.repo.RemoveBook(ctx, orderID, bookID)
}

func (u *orderUseCase) ListAll(ctx context.Context) ([]*domain.Order, error) {
	return u.repo.ListAll(ctx)
}

func (u *orderUseCase) ListByStatus(ctx context.Context, status string) ([]*domain.Order, error) {
	return u.repo.ListByStatus(ctx, status)
}
