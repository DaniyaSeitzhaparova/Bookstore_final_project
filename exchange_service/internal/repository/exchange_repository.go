package repository

import (
	"context"

	"github.com/OshakbayAigerim/read_space/exchange_service/internal/domain"
)

type ExchangeRepository interface {
	CreateOffer(ctx context.Context, offer *domain.ExchangeOffer) (*domain.ExchangeOffer, error)
	GetOffer(ctx context.Context, id string) (*domain.ExchangeOffer, error)
	ListOffersByUser(ctx context.Context, ownerID string) ([]*domain.ExchangeOffer, error)
	ListPendingOffers(ctx context.Context) ([]*domain.ExchangeOffer, error)
	AcceptOffer(ctx context.Context, id string) (*domain.ExchangeOffer, error)
	DeclineOffer(ctx context.Context, id string) (*domain.ExchangeOffer, error)
	DeleteOffer(ctx context.Context, id string) error

	UpdateOffer(ctx context.Context, offer *domain.ExchangeOffer) (*domain.ExchangeOffer, error)
	AddOfferedBook(ctx context.Context, offerID, bookID string) (*domain.ExchangeOffer, error)
	RemoveOfferedBook(ctx context.Context, offerID, bookID string) (*domain.ExchangeOffer, error)
	ListAllOffers(ctx context.Context) ([]*domain.ExchangeOffer, error)
	ListOffersByStatus(ctx context.Context, status string) ([]*domain.ExchangeOffer, error)
}
