package repository

import (
	"context"
	"errors"
	"github.com/OshakbayAigerim/read_space/book_service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoBookRepo struct {
	collection *mongo.Collection
}

func NewMongoBookRepository(client *mongo.Client) *mongoBookRepo {
	return &mongoBookRepo{
		collection: client.Database("readspace").Collection("books"),
	}
}

func (r *mongoBookRepo) Create(ctx context.Context, book *domain.Book) (*domain.Book, error) {
	if book.ID == primitive.NilObjectID {
		book.ID = primitive.NewObjectID()
	}
	_, err := r.collection.InsertOne(ctx, book)
	if err != nil {
		return nil, err
	}
	return book, nil
}

func (r *mongoBookRepo) GetByID(ctx context.Context, id string) (*domain.Book, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid id format")
	}

	var book domain.Book
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&book)
	if err != nil {
		return nil, err
	}
	return &book, nil
}

func (r *mongoBookRepo) ListAll(ctx context.Context) ([]*domain.Book, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var books []*domain.Book
	for cursor.Next(ctx) {
		var book domain.Book
		if err := cursor.Decode(&book); err != nil {
			return nil, err
		}
		books = append(books, &book)
	}
	return books, nil
}

func (r *mongoBookRepo) Update(ctx context.Context, book *domain.Book) (*domain.Book, error) {
	if book.ID == primitive.NilObjectID {
		return nil, errors.New("book ID is empty")
	}
	filter := bson.M{"_id": book.ID}
	update := bson.M{"$set": book}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedBook domain.Book

	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedBook)
	if err != nil {
		return nil, err
	}
	return &updatedBook, nil
}

func (r *mongoBookRepo) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid id format")
	}
	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

func (r *mongoBookRepo) ListByGenre(ctx context.Context, genre string) ([]*domain.Book, error) {
	filter := bson.M{"genre": genre}
	return r.findByFilter(ctx, filter)
}

func (r *mongoBookRepo) ListByAuthor(ctx context.Context, author string) ([]*domain.Book, error) {
	filter := bson.M{"author": author}
	return r.findByFilter(ctx, filter)
}

func (r *mongoBookRepo) ListByLanguage(ctx context.Context, language string) ([]*domain.Book, error) {
	filter := bson.M{"language": language}
	return r.findByFilter(ctx, filter)
}

func (r *mongoBookRepo) ListTopRated(ctx context.Context) ([]*domain.Book, error) {
	opts := options.Find().SetSort(bson.M{"rating": -1}).SetLimit(10)
	return r.findByFilterWithOpts(ctx, bson.M{}, opts)
}

func (r *mongoBookRepo) ListNewArrivals(ctx context.Context) ([]*domain.Book, error) {
	opts := options.Find().SetSort(bson.M{"created_at": -1}).SetLimit(10)
	return r.findByFilterWithOpts(ctx, bson.M{}, opts)
}

func (r *mongoBookRepo) SearchBooks(ctx context.Context, keyword string) ([]*domain.Book, error) {
	filter := bson.M{"$text": bson.M{"$search": keyword}}
	return r.findByFilter(ctx, filter)
}

func (r *mongoBookRepo) RecommendBooks(ctx context.Context, bookID string) ([]*domain.Book, error) {
	return r.ListTopRated(ctx)
}

func (r *mongoBookRepo) findByFilter(ctx context.Context, filter interface{}) ([]*domain.Book, error) {
	return r.findByFilterWithOpts(ctx, filter, nil)
}

func (r *mongoBookRepo) findByFilterWithOpts(ctx context.Context, filter interface{}, opts *options.FindOptions) ([]*domain.Book, error) {
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var books []*domain.Book
	for cursor.Next(ctx) {
		var book domain.Book
		if err := cursor.Decode(&book); err != nil {
			return nil, err
		}
		books = append(books, &book)
	}
	return books, nil
}
