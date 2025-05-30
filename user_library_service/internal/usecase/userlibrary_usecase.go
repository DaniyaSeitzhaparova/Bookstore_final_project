package usecase

import (
	"context"

	"github.com/OshakbayAigerim/read_space/user_library_service/internal/cache"
	"github.com/OshakbayAigerim/read_space/user_library_service/internal/domain"
	"github.com/OshakbayAigerim/read_space/user_library_service/internal/repository"
)

type UserLibraryUseCase interface {
	AssignBook(ctx context.Context, userID, bookID string) (*domain.UserBook, error)
	UnassignBook(ctx context.Context, userID, bookID string) error
	ListUserBooks(ctx context.Context, userID string) ([]*domain.UserBook, error)
}

type userLibraryUseCase struct {
	repo  repository.UserBookRepo
	cache cache.UserLibraryCache
}

func NewUserLibraryUseCase(repo repository.UserBookRepo, c cache.UserLibraryCache) UserLibraryUseCase {
	return &userLibraryUseCase{repo: repo, cache: c}
}

func (uc *userLibraryUseCase) AssignBook(ctx context.Context, userID, bookID string) (*domain.UserBook, error) {
	entry := &domain.UserBook{ /* заполните UserID, BookID */ }
	assigned, err := uc.repo.AssignBook(ctx, entry)
	if err != nil {
		return nil, err
	}
	_ = uc.cache.Invalidate(ctx, userID)
	return assigned, nil
}

func (uc *userLibraryUseCase) UnassignBook(ctx context.Context, userID, bookID string) error {
	if err := uc.repo.UnassignBook(ctx, userID, bookID); err != nil {
		return err
	}
	_ = uc.cache.Invalidate(ctx, userID)
	return nil
}

func (uc *userLibraryUseCase) ListUserBooks(ctx context.Context, userID string) ([]*domain.UserBook, error) {
	return uc.cache.Get(ctx, userID)
}
