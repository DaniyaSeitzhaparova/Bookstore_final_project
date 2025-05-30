package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Book struct {
	ID            primitive.ObjectID 
	Title         string             
	Author        string             
	Genre         string             
	Language      string             
	Description   string             
	Rating        float32            
	Price         float32           
	Pages         int                
	PublishedDate string             
}
