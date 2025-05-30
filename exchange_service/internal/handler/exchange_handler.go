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

	var offeredOIDs, requestedOIDs []primitive.ObjectID
	for _, id := range req.OfferedBookIds {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid offered_book_id %q", id)
		}
		offeredOIDs = append(offeredOIDs, oid)
	}
	for _, id := range req.RequestedBookIds {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid requested_book_id %q", id)
		}
		requestedOIDs = append(requestedOIDs, oid)
	}

	now := primitive.NewDateTimeFromTime(time.Now())
	offer := &domain.ExchangeOffer{
		OwnerID:          ownerOID,
		CounterpartyID:   cpOID,
		OfferedBookIDs:   offeredOIDs,
		RequestedBookIDs: requestedOIDs,
		Status:           "PENDING",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	created, err := h.uc.CreateOffer(ctx, offer)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create offer: %v", err)
	}

	evt := struct {
		OfferID        string `json:"offer_id"`
		OwnerID        string `json:"owner_id"`
		CounterpartyID string `json:"counterparty_id"`
	}{
		OfferID:        created.ID.Hex(),
		OwnerID:        created.OwnerID.Hex(),
		CounterpartyID: created.CounterpartyID.Hex(),
	}
	if data, _ := json.Marshal(evt); data != nil {
		h.nc.Publish("exchange.offered", data)
	}

	toHexs := func(ids []primitive.ObjectID) []string {
		res := make([]string, len(ids))
		for i, x := range ids {
			res[i] = x.Hex()
		}
		return res
	}

	return &exchangepb.OfferResponse{
		Offer: &exchangepb.ExchangeOffer{
			Id:               created.ID.Hex(),
			OwnerId:          created.OwnerID.Hex(),
			CounterpartyId:   created.CounterpartyID.Hex(),
			OfferedBookIds:   toHexs(created.OfferedBookIDs),
			RequestedBookIds: toHexs(created.RequestedBookIDs),
			Status:           created.Status,
			CreatedAt:        created.CreatedAt.Time().String(),
			UpdatedAt:        created.UpdatedAt.Time().String(),
		},
	}, nil
}

func (h *ExchangeHandler) GetOffer(ctx context.Context, req *exchangepb.OfferID) (*exchangepb.OfferResponse, error) {
	if req == nil || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "offer id is required")
	}
	offer, err := h.uc.GetOfferByID(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "offer not found: %v", err)
	}

	toHexs := func(ids []primitive.ObjectID) []string {
		res := make([]string, len(ids))
		for i, x := range ids {
			res[i] = x.Hex()
		}
		return res
	}

	return &exchangepb.OfferResponse{
		Offer: &exchangepb.ExchangeOffer{
			Id:               offer.ID.Hex(),
			OwnerId:          offer.OwnerID.Hex(),
			CounterpartyId:   offer.CounterpartyID.Hex(),
			OfferedBookIds:   toHexs(offer.OfferedBookIDs),
			RequestedBookIds: toHexs(offer.RequestedBookIDs),
			Status:           offer.Status,
			CreatedAt:        offer.CreatedAt.Time().String(),
			UpdatedAt:        offer.UpdatedAt.Time().String(),
		},
	}, nil
}

func (h *ExchangeHandler) ListOffersByUser(ctx context.Context, req *exchangepb.UserID) (*exchangepb.OfferList, error) {
	if req == nil || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	offers, err := h.uc.ListOffersByUser(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot list offers: %v", err)
	}

	var list []*exchangepb.ExchangeOffer
	for _, offer := range offers {
		list = append(list, &exchangepb.ExchangeOffer{
			Id:               offer.ID.Hex(),
			OwnerId:          offer.OwnerID.Hex(),
			CounterpartyId:   offer.CounterpartyID.Hex(),
			OfferedBookIds:   toHexs(offer.OfferedBookIDs),
			RequestedBookIds: toHexs(offer.RequestedBookIDs),
			Status:           offer.Status,
			CreatedAt:        offer.CreatedAt.Time().String(),
			UpdatedAt:        offer.UpdatedAt.Time().String(),
		})
	}

	return &exchangepb.OfferList{Offers: list}, nil
}

func (h *ExchangeHandler) ListPendingOffers(ctx context.Context, _ *exchangepb.Empty) (*exchangepb.OfferList, error) {
	offers, err := h.uc.ListPendingOffers(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot list pending offers: %v", err)
	}

	var list []*exchangepb.ExchangeOffer
	for _, offer := range offers {
		list = append(list, &exchangepb.ExchangeOffer{
			Id:               offer.ID.Hex(),
			OwnerId:          offer.OwnerID.Hex(),
			CounterpartyId:   offer.CounterpartyID.Hex(),
			OfferedBookIds:   toHexs(offer.OfferedBookIDs),
			RequestedBookIds: toHexs(offer.RequestedBookIDs),
			Status:           offer.Status,
			CreatedAt:        offer.CreatedAt.Time().String(),
			UpdatedAt:        offer.UpdatedAt.Time().String(),
		})
	}

	return &exchangepb.OfferList{Offers: list}, nil
}

func (h *ExchangeHandler) AcceptOffer(ctx context.Context, req *exchangepb.AcceptOfferRequest) (*exchangepb.OfferResponse, error) {
	if req == nil || req.OfferId == "" || req.RequesterId == "" {
		return nil, status.Error(codes.InvalidArgument, "offer_id and requester_id are required")
	}

	offer, err := h.uc.AcceptOffer(ctx, req.OfferId, req.RequesterId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot accept offer: %v", err)
	}

	evt := struct {
		OfferID     string `json:"offer_id"`
		OwnerID     string `json:"owner_id"`
		RequesterID string `json:"requester_id"`
	}{
		OfferID:     offer.ID.Hex(),
		OwnerID:     offer.OwnerID.Hex(),
		RequesterID: req.RequesterId,
	}
	if data, _ := json.Marshal(evt); data != nil {
		h.nc.Publish("exchange.accepted", data)
	}

	return &exchangepb.OfferResponse{
		Offer: &exchangepb.ExchangeOffer{
			Id:               offer.ID.Hex(),
			OwnerId:          offer.OwnerID.Hex(),
			CounterpartyId:   offer.CounterpartyID.Hex(),
			OfferedBookIds:   toHexs(offer.OfferedBookIDs),
			RequestedBookIds: toHexs(offer.RequestedBookIDs),
			Status:           offer.Status,
			CreatedAt:        offer.CreatedAt.Time().String(),
			UpdatedAt:        offer.UpdatedAt.Time().String(),
		},
	}, nil
}

func (h *ExchangeHandler) DeclineOffer(ctx context.Context, req *exchangepb.OfferID) (*exchangepb.OfferResponse, error) {
	if req == nil || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "offer id is required")
	}
	o, err := h.uc.DeclineOffer(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot decline offer: %v", err)
	}

	evt := struct {
		OfferID string `json:"offer_id"`
		OwnerID string `json:"owner_id"`
	}{
		OfferID: o.ID.Hex(),
		OwnerID: o.OwnerID.Hex(),
	}
	if data, _ := json.Marshal(evt); data != nil {
		h.nc.Publish("exchange.declined", data)
	}

	return &exchangepb.OfferResponse{
		Offer: &exchangepb.ExchangeOffer{
			Id:               o.ID.Hex(),
			OwnerId:          o.OwnerID.Hex(),
			CounterpartyId:   o.CounterpartyID.Hex(),
			OfferedBookIds:   toHexs(o.OfferedBookIDs),
			RequestedBookIds: toHexs(o.RequestedBookIDs),
			Status:           o.Status,
			CreatedAt:        o.CreatedAt.Time().String(),
			UpdatedAt:        o.UpdatedAt.Time().String(),
		},
	}, nil
}

func (h *ExchangeHandler) DeleteOffer(ctx context.Context, req *exchangepb.OfferID) (*exchangepb.Empty, error) {
	if req == nil || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "offer id is required")
	}
	o, _ := h.uc.GetOfferByID(ctx, req.Id)
	if err := h.uc.DeleteOffer(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "cannot delete offer: %v", err)
	}

	evt := struct {
		OfferID string `json:"offer_id"`
		OwnerID string `json:"owner_id"`
	}{
		OfferID: req.Id,
		OwnerID: o.OwnerID.Hex(),
	}
	if data, _ := json.Marshal(evt); data != nil {
		h.nc.Publish("exchange.deleted", data)
	}

	return &exchangepb.Empty{}, nil
}

func toHexs(ids []primitive.ObjectID) []string {
	out := make([]string, len(ids))
	for i, v := range ids {
		out[i] = v.Hex()
	}
	return out
}
