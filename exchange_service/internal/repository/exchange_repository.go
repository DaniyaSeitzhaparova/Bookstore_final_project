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
}
