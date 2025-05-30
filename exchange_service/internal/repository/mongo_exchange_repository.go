package repository

import (
	"context"
	"time"

	"github.com/OshakbayAigerim/read_space/exchange_service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoExchangeRepo struct {
	collection *mongo.Collection
}

func NewMongoExchangeRepository(db *mongo.Database) ExchangeRepository {
	return &mongoExchangeRepo{
		collection: db.Collection("exchange_offers"),
	}
}

func (r *mongoExchangeRepo) CreateOffer(ctx context.Context, offer *domain.ExchangeOffer) (*domain.ExchangeOffer, error) {
	offer.ID = primitive.NewObjectID()
	now := primitive.NewDateTimeFromTime(time.Now())
	offer.CreatedAt = now
	offer.UpdatedAt = now
	offer.Status = "PENDING"

	if _, err := r.collection.InsertOne(ctx, offer); err != nil {
		return nil, err
	}
	return offer, nil
}

func (r *mongoExchangeRepo) GetOffer(ctx context.Context, id string) (*domain.ExchangeOffer, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var offer domain.ExchangeOffer
	if err := r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&offer); err != nil {
		return nil, err
	}
	return &offer, nil
}

func (r *mongoExchangeRepo) ListOffersByUser(ctx context.Context, ownerID string) ([]*domain.ExchangeOffer, error) {
	oid, err := primitive.ObjectIDFromHex(ownerID)
	if err != nil {
		return nil, err
	}
	cursor, err := r.collection.Find(ctx, bson.M{"owner_id": oid})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var offers []*domain.ExchangeOffer
	for cursor.Next(ctx) {
		var o domain.ExchangeOffer
		if err := cursor.Decode(&o); err != nil {
			return nil, err
		}
		offers = append(offers, &o)
	}
	return offers, nil
}

func (r *mongoExchangeRepo) ListPendingOffers(ctx context.Context) ([]*domain.ExchangeOffer, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"status": "PENDING"})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var offers []*domain.ExchangeOffer
	for cursor.Next(ctx) {
		var o domain.ExchangeOffer
		if err := cursor.Decode(&o); err != nil {
			return nil, err
		}
		offers = append(offers, &o)
	}
	return offers, nil
}

func (r *mongoExchangeRepo) AcceptOffer(ctx context.Context, id string) (*domain.ExchangeOffer, error) {
	return r.updateStatus(ctx, id, "ACCEPTED")
}

func (r *mongoExchangeRepo) DeclineOffer(ctx context.Context, id string) (*domain.ExchangeOffer, error) {
	return r.updateStatus(ctx, id, "DECLINED")
}

func (r *mongoExchangeRepo) updateStatus(ctx context.Context, id, status string) (*domain.ExchangeOffer, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	now := primitive.NewDateTimeFromTime(time.Now())
	after := options.After
	opts := options.FindOneAndUpdateOptions{ReturnDocument: &after}

	filter := bson.M{"_id": objID}
	update := bson.M{"$set": bson.M{
		"status":     status,
		"updated_at": now,
	}}

	var o domain.ExchangeOffer
	if err := r.collection.FindOneAndUpdate(ctx, filter, update, &opts).Decode(&o); err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *mongoExchangeRepo) DeleteOffer(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

// --- New methods ---

func (r *mongoExchangeRepo) UpdateOffer(ctx context.Context, offer *domain.ExchangeOffer) (*domain.ExchangeOffer, error) {
	now := primitive.NewDateTimeFromTime(time.Now())
	after := options.After
	opts := options.FindOneAndUpdateOptions{ReturnDocument: &after}

	filter := bson.M{"_id": offer.ID}
	update := bson.M{"$set": bson.M{
		"owner_id":           offer.OwnerID,
		"counterparty_id":    offer.CounterpartyID,
		"offered_book_ids":   offer.OfferedBookIDs,
		"requested_book_ids": offer.RequestedBookIDs,
		"status":             offer.Status,
		"updated_at":         now,
	}}

	var o domain.ExchangeOffer
	if err := r.collection.FindOneAndUpdate(ctx, filter, update, &opts).Decode(&o); err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *mongoExchangeRepo) AddOfferedBook(ctx context.Context, offerID, bookID string) (*domain.ExchangeOffer, error) {
	objID, err := primitive.ObjectIDFromHex(offerID)
	if err != nil {
		return nil, err
	}
	bid, err := primitive.ObjectIDFromHex(bookID)
	if err != nil {
		return nil, err
	}

	now := primitive.NewDateTimeFromTime(time.Now())
	after := options.After
	opts := options.FindOneAndUpdateOptions{ReturnDocument: &after}

	filter := bson.M{"_id": objID}
	update := bson.M{
		"$push": bson.M{"offered_book_ids": bid},
		"$set":  bson.M{"updated_at": now},
	}

	var o domain.ExchangeOffer
	if err := r.collection.FindOneAndUpdate(ctx, filter, update, &opts).Decode(&o); err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *mongoExchangeRepo) RemoveOfferedBook(ctx context.Context, offerID, bookID string) (*domain.ExchangeOffer, error) {
	objID, err := primitive.ObjectIDFromHex(offerID)
	if err != nil {
		return nil, err
	}
	bid, err := primitive.ObjectIDFromHex(bookID)
	if err != nil {
		return nil, err
	}

	now := primitive.NewDateTimeFromTime(time.Now())
	after := options.After
	opts := options.FindOneAndUpdateOptions{ReturnDocument: &after}

	filter := bson.M{"_id": objID}
	update := bson.M{
		"$pull": bson.M{"offered_book_ids": bid},
		"$set":  bson.M{"updated_at": now},
	}

	var o domain.ExchangeOffer
	if err := r.collection.FindOneAndUpdate(ctx, filter, update, &opts).Decode(&o); err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *mongoExchangeRepo) ListAllOffers(ctx context.Context) ([]*domain.ExchangeOffer, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var offers []*domain.ExchangeOffer
	for cursor.Next(ctx) {
		var o domain.ExchangeOffer
		if err := cursor.Decode(&o); err != nil {
			return nil, err
		}
		offers = append(offers, &o)
	}
	return offers, nil
}

func (r *mongoExchangeRepo) ListOffersByStatus(ctx context.Context, status string) ([]*domain.ExchangeOffer, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"status": status})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var offers []*domain.ExchangeOffer
	for cursor.Next(ctx) {
		var o domain.ExchangeOffer
		if err := cursor.Decode(&o); err != nil {
			return nil, err
		}
		offers = append(offers, &o)
	}
	return offers, nil
}
