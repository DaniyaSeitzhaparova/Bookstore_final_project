package handler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/OshakbayAigerim/read_space/exchange_service/internal/domain"
	"github.com/OshakbayAigerim/read_space/exchange_service/internal/usecase"
	exchangepb "github.com/OshakbayAigerim/read_space/exchange_service/proto"
)

type ExchangeHandler struct {
	exchangepb.UnimplementedExchangeServiceServer
	uc usecase.ExchangeUseCase
	nc *nats.Conn
}

func NewExchangeHandler(uc usecase.ExchangeUseCase, nc *nats.Conn) *ExchangeHandler {
	return &ExchangeHandler{uc: uc, nc: nc}
}

func (h *ExchangeHandler) CreateOffer(ctx context.Context, req *exchangepb.CreateOfferRequest) (*exchangepb.OfferResponse, error) {
	if req == nil || req.OwnerId == "" || req.CounterpartyId == "" ||
		len(req.OfferedBookIds) == 0 || len(req.RequestedBookIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "owner_id, counterparty_id, offered_book_ids and requested_book_ids are required")
	}

	ownerOID, err := primitive.ObjectIDFromHex(req.OwnerId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner_id")
	}
	cpOID, err := primitive.ObjectIDFromHex(req.CounterpartyId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid counterparty_id")
	}

	offered := toObjectIDs(req.OfferedBookIds)
	requested := toObjectIDs(req.RequestedBookIds)

	now := primitive.NewDateTimeFromTime(time.Now())
	offer := &domain.ExchangeOffer{
		OwnerID:          ownerOID,
		CounterpartyID:   cpOID,
		OfferedBookIDs:   offered,
		RequestedBookIDs: requested,
		Status:           "PENDING",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	created, err := h.uc.CreateOffer(ctx, offer)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create offer: %v", err)
	}

	evt := struct {
		OfferID string `json:"offer_id"`
	}{OfferID: created.ID.Hex()}
	if data, _ := json.Marshal(evt); data != nil {
		h.nc.Publish("exchange.created", data)
	}

	return &exchangepb.OfferResponse{Offer: mapDomain(created)}, nil
}

func (h *ExchangeHandler) GetOffer(ctx context.Context, req *exchangepb.OfferID) (*exchangepb.OfferResponse, error) {
	if req == nil || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "offer id is required")
	}
	offer, err := h.uc.GetOfferByID(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "offer not found: %v", err)
	}
	return &exchangepb.OfferResponse{Offer: mapDomain(offer)}, nil
}

func (h *ExchangeHandler) ListOffersByUser(ctx context.Context, req *exchangepb.UserID) (*exchangepb.OfferList, error) {
	if req == nil || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	offers, err := h.uc.ListOffersByUser(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot list offers: %v", err)
	}
	return &exchangepb.OfferList{Offers: mapDomainList(offers)}, nil
}

func (h *ExchangeHandler) ListPendingOffers(ctx context.Context, _ *exchangepb.Empty) (*exchangepb.OfferList, error) {
	offers, err := h.uc.ListPendingOffers(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot list pending offers: %v", err)
	}
	return &exchangepb.OfferList{Offers: mapDomainList(offers)}, nil
}

func (h *ExchangeHandler) AcceptOffer(ctx context.Context, req *exchangepb.AcceptOfferRequest) (*exchangepb.OfferResponse, error) {
	if req == nil || req.OfferId == "" || req.RequesterId == "" {
		return nil, status.Error(codes.InvalidArgument, "offer_id and requester_id are required")
	}
	offer, err := h.uc.AcceptOffer(ctx, req.OfferId, req.RequesterId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot accept offer: %v", err)
	}
	return &exchangepb.OfferResponse{Offer: mapDomain(offer)}, nil
}

func (h *ExchangeHandler) DeclineOffer(ctx context.Context, req *exchangepb.OfferID) (*exchangepb.OfferResponse, error) {
	if req == nil || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "offer id is required")
	}
	offer, err := h.uc.DeclineOffer(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot decline offer: %v", err)
	}
	return &exchangepb.OfferResponse{Offer: mapDomain(offer)}, nil
}

func (h *ExchangeHandler) DeleteOffer(ctx context.Context, req *exchangepb.OfferID) (*exchangepb.Empty, error) {
	if req == nil || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "offer id is required")
	}
	if err := h.uc.DeleteOffer(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "cannot delete offer: %v", err)
	}
	return &exchangepb.Empty{}, nil
}

func (h *ExchangeHandler) UpdateOffer(ctx context.Context, req *exchangepb.UpdateOfferRequest) (*exchangepb.OfferResponse, error) {
	if req == nil || req.Offer == nil || req.Offer.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "offer with id is required")
	}

	oid, err := primitive.ObjectIDFromHex(req.Offer.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid offer id")
	}
	ownerOID, err := primitive.ObjectIDFromHex(req.Offer.OwnerId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner_id")
	}
	cpOID, err := primitive.ObjectIDFromHex(req.Offer.CounterpartyId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid counterparty_id")
	}

	dom := &domain.ExchangeOffer{
		ID:               oid,
		OwnerID:          ownerOID,
		CounterpartyID:   cpOID,
		OfferedBookIDs:   toObjectIDs(req.Offer.OfferedBookIds),
		RequestedBookIDs: toObjectIDs(req.Offer.RequestedBookIds),
		Status:           req.Offer.Status,
		UpdatedAt:        primitive.NewDateTimeFromTime(time.Now()),
	}

	updated, err := h.uc.UpdateOffer(ctx, dom)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot update offer: %v", err)
	}
	return &exchangepb.OfferResponse{Offer: mapDomain(updated)}, nil
}

func (h *ExchangeHandler) AddOfferedBook(ctx context.Context, req *exchangepb.BookOpRequest) (*exchangepb.OfferResponse, error) {
	if req == nil || req.OfferId == "" || req.BookId == "" {
		return nil, status.Error(codes.InvalidArgument, "offer_id and book_id are required")
	}
	offer, err := h.uc.AddOfferedBook(ctx, req.OfferId, req.BookId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot add offered book: %v", err)
	}
	return &exchangepb.OfferResponse{Offer: mapDomain(offer)}, nil
}

func (h *ExchangeHandler) RemoveOfferedBook(ctx context.Context, req *exchangepb.BookOpRequest) (*exchangepb.OfferResponse, error) {
	if req == nil || req.OfferId == "" || req.BookId == "" {
		return nil, status.Error(codes.InvalidArgument, "offer_id and book_id are required")
	}
	offer, err := h.uc.RemoveOfferedBook(ctx, req.OfferId, req.BookId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot remove offered book: %v", err)
	}
	return &exchangepb.OfferResponse{Offer: mapDomain(offer)}, nil
}

func (h *ExchangeHandler) ListAllOffers(ctx context.Context, _ *exchangepb.Empty) (*exchangepb.OfferList, error) {
	offers, err := h.uc.ListAllOffers(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot list all offers: %v", err)
	}
	return &exchangepb.OfferList{Offers: mapDomainList(offers)}, nil
}

func (h *ExchangeHandler) ListOffersByStatus(ctx context.Context, req *exchangepb.StatusRequest) (*exchangepb.OfferList, error) {
	if req == nil || req.Status == "" {
		return nil, status.Error(codes.InvalidArgument, "status is required")
	}
	offers, err := h.uc.ListOffersByStatus(ctx, req.Status)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot list offers by status: %v", err)
	}
	return &exchangepb.OfferList{Offers: mapDomainList(offers)}, nil
}

func mapDomain(o *domain.ExchangeOffer) *exchangepb.ExchangeOffer {
	return &exchangepb.ExchangeOffer{
		Id:               o.ID.Hex(),
		OwnerId:          o.OwnerID.Hex(),
		CounterpartyId:   o.CounterpartyID.Hex(),
		OfferedBookIds:   toHexs(o.OfferedBookIDs),
		RequestedBookIds: toHexs(o.RequestedBookIDs),
		Status:           o.Status,
		CreatedAt:        o.CreatedAt.Time().String(),
		UpdatedAt:        o.UpdatedAt.Time().String(),
	}
}

func mapDomainList(list []*domain.ExchangeOffer) []*exchangepb.ExchangeOffer {
	out := make([]*exchangepb.ExchangeOffer, len(list))
	for i, e := range list {
		out[i] = mapDomain(e)
	}
	return out
}

func toHexs(ids []primitive.ObjectID) []string {
	res := make([]string, len(ids))
	for i, id := range ids {
		res[i] = id.Hex()
	}
	return res
}

func toObjectIDs(strs []string) []primitive.ObjectID {
	out := make([]primitive.ObjectID, len(strs))
	for i, s := range strs {
		if oid, err := primitive.ObjectIDFromHex(s); err == nil {
			out[i] = oid
		}
	}
	return out
}
