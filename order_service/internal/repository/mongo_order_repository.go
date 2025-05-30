package repository

import (
	"context"
	"time"

	"github.com/OshakbayAigerim/read_space/order_service/internal/cache"
	"github.com/OshakbayAigerim/read_space/order_service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoOrderRepo struct {
	collection *mongo.Collection
	cache      cache.OrderCache
}

func NewMongoOrderRepository(db *mongo.Database, cache cache.OrderCache) OrderRepository {
	return &mongoOrderRepo{
		collection: db.Collection("orders"),
		cache:      cache,
	}
}

func (r *mongoOrderRepo) Create(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	order.ID = primitive.NewObjectID()
	now := primitive.NewDateTimeFromTime(time.Now())
	order.CreatedAt = now
	order.UpdatedAt = now

	if _, err := r.collection.InsertOne(ctx, order); err != nil {
		return nil, err
	}
	r.cache.DeleteByUser(ctx, order.UserID.Hex())
	return order, nil
}

func (r *mongoOrderRepo) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	if cached, err := r.cache.Get(ctx, id); err != nil {
		return nil, err
	} else if cached != nil {
		return cached, nil
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var o domain.Order
	if err := r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&o); err != nil {
		return nil, err
	}

	go r.cache.Set(context.Background(), &o)
	return &o, nil
}

func (r *mongoOrderRepo) ListByUser(ctx context.Context, userID string) ([]*domain.Order, error) {
	if cached, err := r.cache.GetByUser(ctx, userID); err != nil {
		return nil, err
	} else if cached != nil {
		return cached, nil
	}

	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": uid})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []*domain.Order
	for cursor.Next(ctx) {
		var o domain.Order
		if err := cursor.Decode(&o); err != nil {
			return nil, err
		}
		orders = append(orders, &o)
	}
	go r.cache.SetByUser(context.Background(), userID, orders)
	return orders, nil
}

func (r *mongoOrderRepo) Cancel(ctx context.Context, id string) (*domain.Order, error) {
	return r.updateStatusWithCache(ctx, id, "Cancelled")
}

func (r *mongoOrderRepo) Return(ctx context.Context, id string) (*domain.Order, error) {
	return r.updateStatusWithCache(ctx, id, "Returned")
}

func (r *mongoOrderRepo) updateStatusWithCache(ctx context.Context, id, status string) (*domain.Order, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var existing domain.Order
	if err := r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&existing); err != nil {
		return nil, err
	}

	now := primitive.NewDateTimeFromTime(time.Now())
	after := options.After
	opt := options.FindOneAndUpdateOptions{ReturnDocument: &after}
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": bson.M{"status": status, "updated_at": now}}

	var updated domain.Order
	if err := r.collection.FindOneAndUpdate(ctx, filter, update, &opt).Decode(&updated); err != nil {
		return nil, err
	}

	r.cache.Delete(ctx, id)
	r.cache.DeleteByUser(ctx, existing.UserID.Hex())
	return &updated, nil
}

func (r *mongoOrderRepo) Update(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	now := primitive.NewDateTimeFromTime(time.Now())
	after := options.After
	opt := options.FindOneAndUpdateOptions{ReturnDocument: &after}
	filter := bson.M{"_id": order.ID}
	update := bson.M{"$set": bson.M{
		"user_id":    order.UserID,
		"book_ids":   order.BookIDs,
		"status":     order.Status,
		"updated_at": now,
	}}

	var updated domain.Order
	if err := r.collection.FindOneAndUpdate(ctx, filter, update, &opt).Decode(&updated); err != nil {
		return nil, err
	}

	r.cache.Delete(ctx, order.ID.Hex())
	r.cache.DeleteByUser(ctx, order.UserID.Hex())
	return &updated, nil
}

func (r *mongoOrderRepo) AddBook(ctx context.Context, orderID, bookID string) (*domain.Order, error) {
	objID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return nil, err
	}
	bid, err := primitive.ObjectIDFromHex(bookID)
	if err != nil {
		return nil, err
	}

	now := primitive.NewDateTimeFromTime(time.Now())
	after := options.After
	opt := options.FindOneAndUpdateOptions{ReturnDocument: &after}
	filter := bson.M{"_id": objID}
	update := bson.M{
		"$push": bson.M{"book_ids": bid},
		"$set":  bson.M{"updated_at": now},
	}

	var updated domain.Order
	if err := r.collection.FindOneAndUpdate(ctx, filter, update, &opt).Decode(&updated); err != nil {
		return nil, err
	}

	r.cache.Delete(ctx, orderID)
	r.cache.DeleteByUser(ctx, updated.UserID.Hex())
	return &updated, nil
}

func (r *mongoOrderRepo) RemoveBook(ctx context.Context, orderID, bookID string) (*domain.Order, error) {
	objID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return nil, err
	}
	bid, err := primitive.ObjectIDFromHex(bookID)
	if err != nil {
		return nil, err
	}

	now := primitive.NewDateTimeFromTime(time.Now())
	after := options.After
	opt := options.FindOneAndUpdateOptions{ReturnDocument: &after}
	filter := bson.M{"_id": objID}
	update := bson.M{
		"$pull": bson.M{"book_ids": bid},
		"$set":  bson.M{"updated_at": now},
	}

	var updated domain.Order
	if err := r.collection.FindOneAndUpdate(ctx, filter, update, &opt).Decode(&updated); err != nil {
		return nil, err
	}

	r.cache.Delete(ctx, orderID)
	r.cache.DeleteByUser(ctx, updated.UserID.Hex())
	return &updated, nil
}

func (r *mongoOrderRepo) ListAll(ctx context.Context) ([]*domain.Order, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []*domain.Order
	for cursor.Next(ctx) {
		var o domain.Order
		if err := cursor.Decode(&o); err != nil {
			return nil, err
		}
		orders = append(orders, &o)
	}
	return orders, nil
}

func (r *mongoOrderRepo) ListByStatus(ctx context.Context, status string) ([]*domain.Order, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"status": status})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []*domain.Order
	for cursor.Next(ctx) {
		var o domain.Order
		if err := cursor.Decode(&o); err != nil {
			return nil, err
		}
		orders = append(orders, &o)
	}
	return orders, nil
}

func (r *mongoOrderRepo) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	var existing domain.Order
	if err := r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&existing); err != nil {
		return err
	}

	if _, err := r.collection.DeleteOne(ctx, bson.M{"_id": objID}); err != nil {
		return err
	}

	r.cache.Delete(ctx, id)
	r.cache.DeleteByUser(ctx, existing.UserID.Hex())
	return nil
}
