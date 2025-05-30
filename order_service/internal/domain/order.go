package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type Order struct {
	ID        primitive.ObjectID
	UserID    primitive.ObjectID
	BookIDs   []primitive.ObjectID
	Status    string
	CreatedAt primitive.DateTime
	UpdatedAt primitive.DateTime
}
