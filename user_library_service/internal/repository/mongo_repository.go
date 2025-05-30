package repository

import (
	"context"
	"github.com/OshakbayAigerim/read_space/user_library_service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoUserBookRepo struct {
	coll *mongo.Collection
}

func NewMongoUserBookRepo(db *mongo.Database) UserBookRepo {
	return &mongoUserBookRepo{coll: db.Collection("user_books")}
}

func (r *mongoUserBookRepo) AssignBook(ctx context.Context, entry *domain.UserBook) (*domain.UserBook, error) {
	entry.ID = primitive.NewObjectID()
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
	cursor, err := r.coll.Find(ctx, bson.M{"user_id": uo})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []*domain.UserBook
	for cursor.Next(ctx) {
		var ub domain.UserBook
		if err := cursor.Decode(&ub); err != nil {
			return nil, err
		}
		list = append(list, &ub)
	}
	return list, nil
}
