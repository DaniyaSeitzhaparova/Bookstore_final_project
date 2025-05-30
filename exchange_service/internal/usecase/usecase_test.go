package usecase

import (
	"context"
	"errors"
	"github.com/OshakbayAigerim/read_space/exchange_service/internal/cache"
	"github.com/OshakbayAigerim/read_space/exchange_service/internal/domain"
	"github.com/OshakbayAigerim/read_space/exchange_service/internal/repository"
	userlibpb "github.com/OshakbayAigerim/read_space/user_library_service/proto"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
)

type fakeRepo struct {
	repository.ExchangeRepository
	acceptCalled, declineCalled, deleteCalled bool
}

func (r *fakeRepo) CreateOffer(ctx context.Context, offer *domain.ExchangeOffer) (*domain.ExchangeOffer, error) {
	return offer, nil
}
func (r *fakeRepo) GetOffer(ctx context.Context, id string) (*domain.ExchangeOffer, error) {
	if id == "err" {
		return nil, errors.New("not found")
	}
	return &domain.ExchangeOffer{ID: primitive.NewObjectID(), OwnerID: primitive.NewObjectID()}, nil
}
func (r *fakeRepo) ListOffersByUser(ctx context.Context, ownerID string) ([]*domain.ExchangeOffer, error) {
	return []*domain.ExchangeOffer{{ID: primitive.NewObjectID()}}, nil
}
func (r *fakeRepo) ListPendingOffers(ctx context.Context) ([]*domain.ExchangeOffer, error) {
	return []*domain.ExchangeOffer{{ID: primitive.NewObjectID()}}, nil
}
func (r *fakeRepo) AcceptOffer(ctx context.Context, id string) (*domain.ExchangeOffer, error) {
	r.acceptCalled = true
	owner := primitive.NewObjectID()
	return &domain.ExchangeOffer{
		ID:               primitive.NewObjectID(),
		OwnerID:          owner,
		OfferedBookIDs:   []primitive.ObjectID{primitive.NewObjectID()},
		RequestedBookIDs: []primitive.ObjectID{primitive.NewObjectID()},
	}, nil
}
func (r *fakeRepo) DeclineOffer(ctx context.Context, id string) (*domain.ExchangeOffer, error) {
	r.declineCalled = true
	return &domain.ExchangeOffer{ID: primitive.NewObjectID(), OwnerID: primitive.NewObjectID()}, nil
}
func (r *fakeRepo) DeleteOffer(ctx context.Context, id string) error {
	r.deleteCalled = true
	return nil
}

type fakeCache struct {
	cache.ExchangeCache
	byUserCalled, pendingCalled bool
	invalUsers                  []string
	invalPending                bool
}

func (c *fakeCache) ListByUser(ctx context.Context, userID string) ([]*domain.ExchangeOffer, error) {
	c.byUserCalled = true
	return []*domain.ExchangeOffer{{ID: primitive.NewObjectID()}}, nil
}
func (c *fakeCache) ListPending(ctx context.Context) ([]*domain.ExchangeOffer, error) {
	c.pendingCalled = true
	return []*domain.ExchangeOffer{{ID: primitive.NewObjectID()}}, nil
}
func (c *fakeCache) InvalidateUser(ctx context.Context, userID string) error {
	c.invalUsers = append(c.invalUsers, userID)
	return nil
}
func (c *fakeCache) InvalidatePending(ctx context.Context) error {
	c.invalPending = true
	return nil
}

type fakeLib struct {
	userlibpb.UserLibraryServiceClient
	unassignErr, assignErr     bool
	unassignCalls, assignCalls int
}

func (f *fakeLib) UnassignBook(ctx context.Context, req *userlibpb.UnassignBookRequest) (*userlibpb.UnassignBookResponse, error) {
	f.unassignCalls++
	if f.unassignErr {
		return nil, errors.New("unassign error")
	}
	return &userlibpb.UnassignBookResponse{Success: true}, nil
}
func (f *fakeLib) AssignBook(ctx context.Context, req *userlibpb.AssignBookRequest) (*userlibpb.AssignBookResponse, error) {
	f.assignCalls++
	if f.assignErr {
		return nil, errors.New("assign error")
	}
	return &userlibpb.AssignBookResponse{Entry: &userlibpb.UserBook{}}, nil
}

func (f *fakeLib) ListUserBooks(ctx context.Context, req *userlibpb.ListUserBooksRequest) (*userlibpb.ListUserBooksResponse, error) {
	return &userlibpb.ListUserBooksResponse{}, nil
}

func TestCreateOffer_InvalidatesCache(t *testing.T) {
	repo := &fakeRepo{}
	cache := &fakeCache{}
	uc := NewExchangeUseCase(repo, cache, nil)

	owner := primitive.NewObjectID()
	off, err := uc.CreateOffer(context.Background(), &domain.ExchangeOffer{OwnerID: owner})
	if err != nil {
		t.Fatal(err)
	}
	if !cache.invalPending {
		t.Error("pending cache not invalidated")
	}
	if len(cache.invalUsers) != 1 || cache.invalUsers[0] != owner.Hex() {
		t.Errorf("user cache not invalidated, got %v", cache.invalUsers)
	}
	if off.OwnerID != owner {
		t.Errorf("expected OwnerID %v, got %v", owner, off.OwnerID)
	}
}

func TestListMethods_UseCache(t *testing.T) {
	uc := NewExchangeUseCase(&fakeRepo{}, &fakeCache{}, nil)

	uc.ListOffersByUser(context.Background(), "u1")
	fc := uc.(*exchangeUseCase).cache.(*fakeCache)
	if !fc.byUserCalled {
		t.Error("ListOffersByUser did not call cache.ListByUser")
	}

	uc.ListPendingOffers(context.Background())
	if !fc.pendingCalled {
		t.Error("ListPendingOffers did not call cache.ListPending")
	}
}

func TestAcceptOffer_FlowAndInvalidate(t *testing.T) {
	repo := &fakeRepo{}
	cache := &fakeCache{}
	lib := &fakeLib{}
	uc := NewExchangeUseCase(repo, cache, lib)

	_, err := uc.AcceptOffer(context.Background(), "id", "req")
	if err != nil {
		t.Fatal(err)
	}
	if !repo.acceptCalled {
		t.Error("expected repo.AcceptOffer")
	}
	if lib.unassignCalls != 2 || lib.assignCalls != 2 {
		t.Errorf("expected 2 unassign & 2 assign, got %d/%d", lib.unassignCalls, lib.assignCalls)
	}
	if !cache.invalPending || len(cache.invalUsers) != 2 {
		t.Errorf("expected pending+2 users invalidated, got %v/%v", cache.invalPending, cache.invalUsers)
	}
}

func TestDeclineAndDelete_Invalidation(t *testing.T) {
	repo := &fakeRepo{}
	cache := &fakeCache{}
	uc := NewExchangeUseCase(repo, cache, nil)

	uc.DeclineOffer(context.Background(), "id")
	if !repo.declineCalled || !cache.invalPending || len(cache.invalUsers) != 1 {
		t.Error("DeclineOffer failed cache/repo calls")
	}

	repo = &fakeRepo{}
	cache = &fakeCache{}
	uc = NewExchangeUseCase(repo, cache, nil)

	uc.DeleteOffer(context.Background(), "id")
	if !repo.deleteCalled || !cache.invalPending || len(cache.invalUsers) != 1 {
		t.Error("DeleteOffer failed cache/repo calls")
	}
}

func TestGetOffer_Error(t *testing.T) {
	repo := &fakeRepo{}
	cache := &fakeCache{}
	uc := NewExchangeUseCase(repo, cache, nil)

	_, err := uc.GetOfferByID(context.Background(), "err")
	if err == nil {
		t.Error("expected error from GetOfferByID")
	}
}
