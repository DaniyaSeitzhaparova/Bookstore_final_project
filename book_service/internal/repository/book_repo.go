// internal/repository/book_repository.go
package repository

import (
	"context"
	"github.com/OshakbayAigerim/read_space/book_service/internal/domain"
)

type BookRepository interface {
	Create(ctx context.Context, book *domain.Book) (*domain.Book, error)
	GetByID(ctx context.Context, id string) (*domain.Book, error)
	ListAll(ctx context.Context) ([]*domain.Book, error)
	Update(ctx context.Context, book *domain.Book) (*domain.Book, error)
	Delete(ctx context.Context, id string) error
	ListByGenre(ctx context.Context, genre string) ([]*domain.Book, error)
	ListByAuthor(ctx context.Context, author string) ([]*domain.Book, error)
	ListByLanguage(ctx context.Context, language string) ([]*domain.Book, error)
	ListTopRated(ctx context.Context) ([]*domain.Book, error)
	ListNewArrivals(ctx context.Context) ([]*domain.Book, error)
	SearchBooks(ctx context.Context, keyword string) ([]*domain.Book, error)
	RecommendBooks(ctx context.Context, bookID string) ([]*domain.Book, error)
}
