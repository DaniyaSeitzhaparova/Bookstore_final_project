package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type ExchangeOffer struct {
	ID               primitive.ObjectID
	OwnerID          primitive.ObjectID
	CounterpartyID   primitive.ObjectID
	OfferedBookIDs   []primitive.ObjectID
	RequestedBookIDs []primitive.ObjectID
	Status           string
	CreatedAt        primitive.DateTime
	UpdatedAt        primitive.DateTime
}
