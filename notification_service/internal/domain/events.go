package domain

type OrderCreatedEvent struct {
	OrderID string   `json:"order_id"`
	UserID  string   `json:"user_id"`
	BookIDs []string `json:"book_ids"`
}

type UserCreatedEvent struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type OrderCompletedEvent struct {
	OrderID string   `json:"order_id"`
	UserID  string   `json:"user_id"`
	BookIDs []string `json:"book_ids"`
}

type OrderDeletedEvent struct {
	OrderID string   `json:"order_id"`
	UserID  string   `json:"user_id"`
	BookIDs []string `json:"book_ids"`
}

type OfferCreatedEvent struct {
	OfferID string `json:"offer_id"`
	OwnerID string `json:"owner_id"`
}
type OfferDeclinedEvent struct {
	OfferID string `json:"offer_id"`
	OwnerID string `json:"owner_id"`
}

type OfferAcceptedEvent struct {
	OfferID      string `json:"offer_id"`
	OwnerID      string `json:"owner_id"`
	Counterparty string `json:"requester_id"`
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
