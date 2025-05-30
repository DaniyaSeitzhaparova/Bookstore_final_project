package repository

import (
	"context"
	"log"
	"time"

	"github.com/OshakbayAigerim/read_space/book_service/internal/cache"
	"github.com/OshakbayAigerim/read_space/book_service/internal/domain"
)

type cachedBookRepo struct {
	repo  BookRepository
	cache cache.BookCache
}

func NewCachedBookRepository(repo BookRepository, cache cache.BookCache) BookRepository {
	return &cachedBookRepo{
		repo:  repo,
		cache: cache,
	}
}

func (r *cachedBookRepo) getCacheKeyForBook(id string) string {
	return "book:" + id
}

func (r *cachedBookRepo) getCacheKeyForList(listType string) string {
	return "books:" + listType
}

func (r *cachedBookRepo) Create(ctx context.Context, book *domain.Book) (*domain.Book, error) {
	created, err := r.repo.Create(ctx, book)
	if err != nil {
		return nil, err
	}
	r.cache.Delete(ctx, r.getCacheKeyForList("all"))
	return created, nil
}

func (r *cachedBookRepo) GetByID(ctx context.Context, id string) (*domain.Book, error) {
	cacheKey := r.getCacheKeyForBook(id)

	book, err := r.cache.Get(ctx, cacheKey)
	if err == nil {
		log.Println("cache hit for key:", cacheKey)
		return book, nil
	}

	log.Println("cache miss for key:", cacheKey)

	book, err = r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	_ = r.cache.Set(ctx, cacheKey, book, 10*time.Minute)
	return book, nil
}

func (r *cachedBookRepo) ListAll(ctx context.Context) ([]*domain.Book, error) {
	cacheKey := r.getCacheKeyForList("all")
	if books, err := r.cache.GetList(ctx, cacheKey); err == nil && books != nil {
		return books, nil
	}

	books, err := r.repo.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	r.cache.SetList(ctx, cacheKey, books, 5*time.Minute)
	return books, nil
}

func (r *cachedBookRepo) Update(ctx context.Context, book *domain.Book) (*domain.Book, error) {
	updated, err := r.repo.Update(ctx, book)
	if err != nil {
		return nil, err
	}
	r.cache.Delete(ctx, r.getCacheKeyForBook(updated.ID.Hex()))
	r.cache.Delete(ctx, r.getCacheKeyForList("all"))
	return updated, nil
}

func (r *cachedBookRepo) Delete(ctx context.Context, id string) error {
	err := r.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	r.cache.Delete(ctx, r.getCacheKeyForBook(id))
	r.cache.Delete(ctx, r.getCacheKeyForList("all"))
	return nil
}

func (r *cachedBookRepo) ListByGenre(ctx context.Context, genre string) ([]*domain.Book, error) {
	cacheKey := r.getCacheKeyForList("genre:" + genre)
	if books, err := r.cache.GetList(ctx, cacheKey); err == nil && books != nil {
		return books, nil
	}
	books, err := r.repo.ListByGenre(ctx, genre)
	if err != nil {
		return nil, err
	}
	r.cache.SetList(ctx, cacheKey, books, 10*time.Minute)
	return books, nil
}

func (r *cachedBookRepo) ListByAuthor(ctx context.Context, author string) ([]*domain.Book, error) {
	cacheKey := r.getCacheKeyForList("author:" + author)
	if books, err := r.cache.GetList(ctx, cacheKey); err == nil && books != nil {
		return books, nil
	}
	books, err := r.repo.ListByAuthor(ctx, author)
	if err != nil {
		return nil, err
	}
	r.cache.SetList(ctx, cacheKey, books, 10*time.Minute)
	return books, nil
}

func (r *cachedBookRepo) ListByLanguage(ctx context.Context, language string) ([]*domain.Book, error) {
	cacheKey := r.getCacheKeyForList("language:" + language)
	if books, err := r.cache.GetList(ctx, cacheKey); err == nil && books != nil {
		return books, nil
	}
	books, err := r.repo.ListByLanguage(ctx, language)
	if err != nil {
		return nil, err
	}
	r.cache.SetList(ctx, cacheKey, books, 10*time.Minute)
	return books, nil
}

func (r *cachedBookRepo) ListTopRated(ctx context.Context) ([]*domain.Book, error) {
	cacheKey := r.getCacheKeyForList("top_rated")
	if books, err := r.cache.GetList(ctx, cacheKey); err == nil && books != nil {
		return books, nil
	}
	books, err := r.repo.ListTopRated(ctx)
	if err != nil {
		return nil, err
	}
	r.cache.SetList(ctx, cacheKey, books, 15*time.Minute)
	return books, nil
}

func (r *cachedBookRepo) ListNewArrivals(ctx context.Context) ([]*domain.Book, error) {
	cacheKey := r.getCacheKeyForList("new_arrivals")
	if books, err := r.cache.GetList(ctx, cacheKey); err == nil && books != nil {
		return books, nil
	}
	books, err := r.repo.ListNewArrivals(ctx)
	if err != nil {
		return nil, err
	}
	r.cache.SetList(ctx, cacheKey, books, 30*time.Minute)
	return books, nil
}

func (r *cachedBookRepo) SearchBooks(ctx context.Context, keyword string) ([]*domain.Book, error) {
	return r.repo.SearchBooks(ctx, keyword)
}

func (r *cachedBookRepo) RecommendBooks(ctx context.Context, bookID string) ([]*domain.Book, error) {
	cacheKey := r.getCacheKeyForList("recommend:" + bookID)
	if books, err := r.cache.GetList(ctx, cacheKey); err == nil && books != nil {
		return books, nil
	}
	books, err := r.repo.RecommendBooks(ctx, bookID)
	if err != nil {
		return nil, err
	}
	r.cache.SetList(ctx, cacheKey, books, 30*time.Minute)
	return books, nil
}
