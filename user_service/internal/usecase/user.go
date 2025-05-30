package usecase

import (
	"context"

	"github.com/OshakbayAigerim/read_space/user_service/internal/domain"
	"github.com/OshakbayAigerim/read_space/user_service/internal/repository"
)

type UserUseCase interface {
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	ListUsers(ctx context.Context) ([]*domain.User, error)
}

type userUseCase struct {
	repo repository.UserRepository
}

func NewUserUseCase(r repository.UserRepository) UserUseCase {
	return &userUseCase{repo: r}
}

func (u *userUseCase) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	return u.repo.Create(ctx, user)
}

func (u *userUseCase) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *userUseCase) ListUsers(ctx context.Context) ([]*domain.User, error) {
	return u.repo.ListAll(ctx)
}
