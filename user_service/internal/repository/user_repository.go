package repository

import (
	"context"

	"github.com/OshakbayAigerim/read_space/user_service/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	ListAll(ctx context.Context) ([]*domain.User, error)
}
