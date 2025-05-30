package usecase

import (
	"context"

	"github.com/OshakbayAigerim/read_space/exchange_service/internal/cache"
	"github.com/OshakbayAigerim/read_space/exchange_service/internal/domain"
	"github.com/OshakbayAigerim/read_space/exchange_service/internal/repository"
	userlibpb "github.com/OshakbayAigerim/read_space/user_library_service/proto"
)

type ExchangeUseCase interface {
	CreateOffer(ctx context.Context, offer *domain.ExchangeOffer) (*domain.ExchangeOffer, error)
	GetOfferByID(ctx context.Context, id string) (*domain.ExchangeOffer, error)
	ListOffersByUser(ctx context.Context, ownerID string) ([]*domain.ExchangeOffer, error)
	ListPendingOffers(ctx context.Context) ([]*domain.ExchangeOffer, error)
	AcceptOffer(ctx context.Context, offerID, requesterID string) (*domain.ExchangeOffer, error)
	DeclineOffer(ctx context.Context, id string) (*domain.ExchangeOffer, error)
	DeleteOffer(ctx context.Context, id string) error

	UpdateOffer(ctx context.Context, offer *domain.ExchangeOffer) (*domain.ExchangeOffer, error)
	AddOfferedBook(ctx context.Context, offerID, bookID string) (*domain.ExchangeOffer, error)
	RemoveOfferedBook(ctx context.Context, offerID, bookID string) (*domain.ExchangeOffer, error)
	ListAllOffers(ctx context.Context) ([]*domain.ExchangeOffer, error)
	ListOffersByStatus(ctx context.Context, status string) ([]*domain.ExchangeOffer, error)
}

type exchangeUseCase struct {
	repo      repository.ExchangeRepository
	cache     cache.ExchangeCache
	libClient userlibpb.UserLibraryServiceClient
}

func NewExchangeUseCase(
	r repository.ExchangeRepository,
	c cache.ExchangeCache,
	lc userlibpb.UserLibraryServiceClient,
) ExchangeUseCase {
	return &exchangeUseCase{
		repo:      r,
		cache:     c,
		libClient: lc,
	}
}

func (u *exchangeUseCase) CreateOffer(ctx context.Context, offer *domain.ExchangeOffer) (*domain.ExchangeOffer, error) {
	created, err := u.repo.CreateOffer(ctx, offer)
	if err != nil {
		return nil, err
	}
	u.cache.InvalidatePending(ctx)
	u.cache.InvalidateUser(ctx, offer.OwnerID.Hex())
	return created, nil
}

func (u *exchangeUseCase) GetOfferByID(ctx context.Context, id string) (*domain.ExchangeOffer, error) {
	return u.repo.GetOffer(ctx, id)
}

func (u *exchangeUseCase) ListOffersByUser(ctx context.Context, ownerID string) ([]*domain.ExchangeOffer, error) {
	return u.cache.ListByUser(ctx, ownerID)
}

func (u *exchangeUseCase) ListPendingOffers(ctx context.Context) ([]*domain.ExchangeOffer, error) {
	return u.cache.ListPending(ctx)
}

func (u *exchangeUseCase) AcceptOffer(ctx context.Context, offerID, requesterID string) (*domain.ExchangeOffer, error) {
	offer, err := u.repo.AcceptOffer(ctx, offerID)
	if err != nil {
		return nil, err
	}
	u.cache.InvalidatePending(ctx)
	u.cache.InvalidateUser(ctx, offer.OwnerID.Hex())
	u.cache.InvalidateUser(ctx, requesterID)
	return offer, nil
}

func (u *exchangeUseCase) DeclineOffer(ctx context.Context, id string) (*domain.ExchangeOffer, error) {
	o, err := u.repo.DeclineOffer(ctx, id)
	if err != nil {
		return nil, err
	}
	u.cache.InvalidatePending(ctx)
	u.cache.InvalidateUser(ctx, o.OwnerID.Hex())
	return o, nil
}

func (u *exchangeUseCase) DeleteOffer(ctx context.Context, id string) error {
	o, err := u.repo.GetOffer(ctx, id)
	if err != nil {
		return err
	}
	if err := u.repo.DeleteOffer(ctx, id); err != nil {
		return err
	}
	u.cache.InvalidatePending(ctx)
	u.cache.InvalidateUser(ctx, o.OwnerID.Hex())
	return nil
}

func (u *exchangeUseCase) UpdateOffer(ctx context.Context, offer *domain.ExchangeOffer) (*domain.ExchangeOffer, error) {
	updated, err := u.repo.UpdateOffer(ctx, offer)
	if err != nil {
		return nil, err
	}
	u.cache.InvalidatePending(ctx)
	u.cache.InvalidateUser(ctx, offer.OwnerID.Hex())
	return updated, nil
}

func (u *exchangeUseCase) AddOfferedBook(ctx context.Context, offerID, bookID string) (*domain.ExchangeOffer, error) {
	updated, err := u.repo.AddOfferedBook(ctx, offerID, bookID)
	if err != nil {
		return nil, err
	}
	u.cache.InvalidatePending(ctx)
	u.cache.InvalidateUser(ctx, updated.OwnerID.Hex())
	return updated, nil
}

func (u *exchangeUseCase) RemoveOfferedBook(ctx context.Context, offerID, bookID string) (*domain.ExchangeOffer, error) {
	updated, err := u.repo.RemoveOfferedBook(ctx, offerID, bookID)
	if err != nil {
		return nil, err
	}
	u.cache.InvalidatePending(ctx)
	u.cache.InvalidateUser(ctx, updated.OwnerID.Hex())
	return updated, nil
}

func (u *exchangeUseCase) ListAllOffers(ctx context.Context) ([]*domain.ExchangeOffer, error) {
	return u.repo.ListAllOffers(ctx)
}

func (u *exchangeUseCase) ListOffersByStatus(ctx context.Context, status string) ([]*domain.ExchangeOffer, error) {
	return u.repo.ListOffersByStatus(ctx, status)
}
