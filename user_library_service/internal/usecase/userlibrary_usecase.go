package usecase

import (
	"context"

	"github.com/OshakbayAigerim/read_space/user_library_service/internal/cache"
	"github.com/OshakbayAigerim/read_space/user_library_service/internal/domain"
	"github.com/OshakbayAigerim/read_space/user_library_service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserLibraryUseCase interface {
	AssignBook(ctx context.Context, userID, bookID string) (*domain.UserBook, error)
	UnassignBook(ctx context.Context, userID, bookID string) error
	ListUserBooks(ctx context.Context, userID string) ([]*domain.UserBook, error)

	GetEntry(ctx context.Context, id string) (*domain.UserBook, error)
	DeleteEntry(ctx context.Context, id string) error
	UpdateEntry(ctx context.Context, ub *domain.UserBook) (*domain.UserBook, error)
	ListAllEntries(ctx context.Context) ([]*domain.UserBook, error)
	ListByBook(ctx context.Context, bookID string) ([]*domain.UserBook, error)
}

type userLibraryUseCase struct {
	repo  repository.UserBookRepo
	cache cache.UserLibraryCache
}

func NewUserLibraryUseCase(repo repository.UserBookRepo, c cache.UserLibraryCache) UserLibraryUseCase {
	return &userLibraryUseCase{repo: repo, cache: c}
}

func (uc *userLibraryUseCase) AssignBook(ctx context.Context, userID, bookID string) (*domain.UserBook, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}
	bid, err := primitive.ObjectIDFromHex(bookID)
	if err != nil {
		return nil, err
	}

	entry := &domain.UserBook{
		ID:     primitive.NewObjectID(),
		UserID: uid,
		BookID: bid,
	}

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

func (uc *userLibraryUseCase) GetEntry(ctx context.Context, id string) (*domain.UserBook, error) {
	return uc.repo.GetEntry(ctx, id)
}

func (uc *userLibraryUseCase) DeleteEntry(ctx context.Context, id string) error {
	return uc.repo.DeleteEntry(ctx, id)
}

func (uc *userLibraryUseCase) UpdateEntry(ctx context.Context, ub *domain.UserBook) (*domain.UserBook, error) {
	updated, err := uc.repo.UpdateEntry(ctx, ub)
	if err != nil {
		return nil, err
	}
	_ = uc.cache.Invalidate(ctx, ub.UserID.Hex())
	return updated, nil
}

func (uc *userLibraryUseCase) ListAllEntries(ctx context.Context) ([]*domain.UserBook, error) {
	return uc.repo.ListAllEntries(ctx)
}

func (uc *userLibraryUseCase) ListByBook(ctx context.Context, bookID string) ([]*domain.UserBook, error) {
	return uc.repo.ListByBook(ctx, bookID)
}
