package usecase

import (
	"context"
	"fmt"

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
	_ = u.cache.InvalidatePending(ctx)
	_ = u.cache.InvalidateUser(ctx, offer.OwnerID.Hex())
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

	for _, bID := range offer.OfferedBookIDs {
		if _, err := u.libClient.UnassignBook(ctx, &userlibpb.UnassignBookRequest{
			UserId: offer.OwnerID.Hex(),
			BookId: bID.Hex(),
		}); err != nil {
			u.repo.DeclineOffer(ctx, offerID)
			return nil, fmt.Errorf("unassign offered book %s: %w", bID.Hex(), err)
		}
		if _, err := u.libClient.AssignBook(ctx, &userlibpb.AssignBookRequest{
			UserId: requesterID,
			BookId: bID.Hex(),
		}); err != nil {
			u.repo.DeclineOffer(ctx, offerID)
			return nil, fmt.Errorf("assign offered book %s: %w", bID.Hex(), err)
		}
	}

	for _, bID := range offer.RequestedBookIDs {
		if _, err := u.libClient.UnassignBook(ctx, &userlibpb.UnassignBookRequest{
			UserId: requesterID,
			BookId: bID.Hex(),
		}); err != nil {
			u.repo.DeclineOffer(ctx, offerID)
			return nil, fmt.Errorf("unassign requested book %s: %w", bID.Hex(), err)
		}
		if _, err := u.libClient.AssignBook(ctx, &userlibpb.AssignBookRequest{
			UserId: offer.OwnerID.Hex(),
			BookId: bID.Hex(),
		}); err != nil {
			u.repo.DeclineOffer(ctx, offerID)
			return nil, fmt.Errorf("assign requested book %s: %w", bID.Hex(), err)
		}
	}

	_ = u.cache.InvalidatePending(ctx)
	_ = u.cache.InvalidateUser(ctx, offer.OwnerID.Hex())
	_ = u.cache.InvalidateUser(ctx, requesterID)

	return offer, nil
}

func (u *exchangeUseCase) DeclineOffer(ctx context.Context, id string) (*domain.ExchangeOffer, error) {
	o, err := u.repo.DeclineOffer(ctx, id)
	if err != nil {
		return nil, err
	}
	_ = u.cache.InvalidatePending(ctx)
	_ = u.cache.InvalidateUser(ctx, o.OwnerID.Hex())
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
	_ = u.cache.InvalidatePending(ctx)
	_ = u.cache.InvalidateUser(ctx, o.OwnerID.Hex())
	return nil
}
