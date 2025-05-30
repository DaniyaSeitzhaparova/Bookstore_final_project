package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserBook struct {
	ID     primitive.ObjectID
	UserID primitive.ObjectID
	BookID primitive.ObjectID
}

type BookAssignedEvent struct {
	UserID string `json:"user_id"`
	BookID string `json:"book_id"`
}

type BookUnassignedEvent struct {
	UserID string `json:"user_id"`
	BookID string `json:"book_id"`
}

type EntryDeletedEvent struct {
	EntryID string `json:"id"`
	UserID  string `json:"user_id"`
}

type EntryUpdatedEvent struct {
	EntryID string `json:"id"`
	UserID  string `json:"user_id"`
	BookID  string `json:"book_id"`
}
