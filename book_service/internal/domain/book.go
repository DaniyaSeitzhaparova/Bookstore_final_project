package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Book struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Title         string             `bson:"title"`
	Author        string             `bson:"author"`
	Genre         string             `bson:"genre"`
	Language      string             `bson:"language"`
	Description   string             `bson:"description"`
	Rating        float32            `bson:"rating"`
	Price         float32            `bson:"price"`
	Pages         int                `bson:"pages"`
	PublishedDate string             `bson:"published_date"`
}
