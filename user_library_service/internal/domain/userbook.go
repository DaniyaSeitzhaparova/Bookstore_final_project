package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserBook struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	UserID primitive.ObjectID `bson:"user_id"`
	BookID primitive.ObjectID `bson:"book_id"`
}
