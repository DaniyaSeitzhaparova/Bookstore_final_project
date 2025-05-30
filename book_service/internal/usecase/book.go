package usecase

import (
	"context"
	"github.com/OshakbayAigerim/read_space/book_service/internal/domain"
	"github.com/OshakbayAigerim/read_space/book_service/internal/repository"
)

type BookUseCase interface {
	CreateBook(ctx context.Context, book *domain.Book) (*domain.Book, error)
	GetBookByID(ctx context.Context, id string) (*domain.Book, error)
	ListBooks(ctx context.Context) ([]*domain.Book, error)
	UpdateBook(ctx context.Context, book *domain.Book) (*domain.Book, error)
	DeleteBook(ctx context.Context, id string) error
	ListBooksByGenre(ctx context.Context, genre string) ([]*domain.Book, error)
	ListBooksByAuthor(ctx context.Context, author string) ([]*domain.Book, error)
	ListBooksByLanguage(ctx context.Context, language string) ([]*domain.Book, error)
	ListTopRated(ctx context.Context) ([]*domain.Book, error)
	ListNewArrivals(ctx context.Context) ([]*domain.Book, error)
	SearchBooks(ctx context.Context, keyword string) ([]*domain.Book, error)
	RecommendBooks(ctx context.Context, bookID string) ([]*domain.Book, error)
}

type bookUseCase struct {
	repo repository.BookRepository
}

func NewBookUseCase(r repository.BookRepository) BookUseCase {
	return &bookUseCase{
		repo: r,
	}
}

func (u *bookUseCase) CreateBook(ctx context.Context, book *domain.Book) (*domain.Book, error) {
	return u.repo.Create(ctx, book)
}

func (u *bookUseCase) GetBookByID(ctx context.Context, id string) (*domain.Book, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *bookUseCase) ListBooks(ctx context.Context) ([]*domain.Book, error) {
	return u.repo.ListAll(ctx)
}

func (u *bookUseCase) UpdateBook(ctx context.Context, book *domain.Book) (*domain.Book, error) {
	return u.repo.Update(ctx, book)
}

func (u *bookUseCase) DeleteBook(ctx context.Context, id string) error {
	return u.repo.Delete(ctx, id)
}

func (u *bookUseCase) ListBooksByGenre(ctx context.Context, genre string) ([]*domain.Book, error) {
	return u.repo.ListByGenre(ctx, genre)
}

func (u *bookUseCase) ListBooksByAuthor(ctx context.Context, author string) ([]*domain.Book, error) {
	return u.repo.ListByAuthor(ctx, author)
}

func (u *bookUseCase) ListBooksByLanguage(ctx context.Context, language string) ([]*domain.Book, error) {
	return u.repo.ListByLanguage(ctx, language)
}

func (u *bookUseCase) ListTopRated(ctx context.Context) ([]*domain.Book, error) {
	return u.repo.ListTopRated(ctx)
}

func (u *bookUseCase) ListNewArrivals(ctx context.Context) ([]*domain.Book, error) {
	return u.repo.ListNewArrivals(ctx)
}

func (u *bookUseCase) SearchBooks(ctx context.Context, keyword string) ([]*domain.Book, error) {
	return u.repo.SearchBooks(ctx, keyword)
}

func (u *bookUseCase) RecommendBooks(ctx context.Context, bookID string) ([]*domain.Book, error) {
	return u.repo.RecommendBooks(ctx, bookID)
}
