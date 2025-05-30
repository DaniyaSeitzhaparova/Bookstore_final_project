package repository

import (
	"context"

	"github.com/OshakbayAigerim/read_space/user_library_service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoUserBookRepo struct {
	coll *mongo.Collection
}

func NewMongoUserBookRepo(db *mongo.Database) UserBookRepo {
	return &mongoUserBookRepo{coll: db.Collection("user_books")}
}

func (r *mongoUserBookRepo) AssignBook(ctx context.Context, entry *domain.UserBook) (*domain.UserBook, error) {
	entry.ID = primitive.NewObjectID()
	// insert into Mongo
	if _, err := r.coll.InsertOne(ctx, entry); err != nil {
		return nil, err
	}
	return entry, nil
}

func (r *mongoUserBookRepo) UnassignBook(ctx context.Context, userID, bookID string) error {
	uo, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	bo, err := primitive.ObjectIDFromHex(bookID)
	if err != nil {
		return err
	}
	_, err = r.coll.DeleteOne(ctx, bson.M{"user_id": uo, "book_id": bo})
	return err
}

func (r *mongoUserBookRepo) ListUserBooks(ctx context.Context, userID string) ([]*domain.UserBook, error) {
	uo, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}
	cur, err := r.coll.Find(ctx, bson.M{"user_id": uo})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var out []*domain.UserBook
	for cur.Next(ctx) {
		var u domain.UserBook
		if err := cur.Decode(&u); err != nil {
			return nil, err
		}
		out = append(out, &u)
	}
	return out, nil
}

func (r *mongoUserBookRepo) GetEntry(ctx context.Context, id string) (*domain.UserBook, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var u domain.UserBook
	if err := r.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *mongoUserBookRepo) DeleteEntry(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.coll.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}

func (r *mongoUserBookRepo) UpdateEntry(ctx context.Context, entry *domain.UserBook) (*domain.UserBook, error) {
	// only user_id and book_id can change
	update := bson.M{
		"$set": bson.M{
			"user_id": entry.UserID,
			"book_id": entry.BookID,
		},
	}
	after := options.After
	opt := options.FindOneAndUpdateOptions{ReturnDocument: &after}

	var updated domain.UserBook
	if err := r.coll.
		FindOneAndUpdate(ctx, bson.M{"_id": entry.ID}, update, &opt).
		Decode(&updated); err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *mongoUserBookRepo) ListAllEntries(ctx context.Context) ([]*domain.UserBook, error) {
	cur, err := r.coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var out []*domain.UserBook
	for cur.Next(ctx) {
		var u domain.UserBook
		if err := cur.Decode(&u); err != nil {
			return nil, err
		}
		out = append(out, &u)
	}
	return out, nil
}

func (r *mongoUserBookRepo) ListByBook(ctx context.Context, bookID string) ([]*domain.UserBook, error) {
	bo, err := primitive.ObjectIDFromHex(bookID)
	if err != nil {
		return nil, err
	}
	cur, err := r.coll.Find(ctx, bson.M{"book_id": bo})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var out []*domain.UserBook
	for cur.Next(ctx) {
		var u domain.UserBook
		if err := cur.Decode(&u); err != nil {
			return nil, err
		}
		out = append(out, &u)
	}
	return out, nil
}
