package repository

import (
	"context"
	"time"

	"github.com/OshakbayAigerim/read_space/user_service/internal/cache"
	"github.com/OshakbayAigerim/read_space/user_service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoUserRepo struct {
	collection *mongo.Collection
	cache      *cache.UserCache
}

func NewMongoUserRepository(db *mongo.Database, cache *cache.UserCache) UserRepository {
	return &mongoUserRepo{
		collection: db.Collection("users"),
		cache:      cache,
	}
}

func (r *mongoUserRepo) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	user.ID = primitive.NewObjectID()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}

	go func() {
		_ = r.cache.Set(context.Background(), user)
	}()

	return user, nil
}

func (r *mongoUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if user, err := r.cache.Get(ctx, id); err == nil && user != nil {
		return user, nil
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var user domain.User
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	go func() {
		_ = r.cache.Set(context.Background(), &user)
	}()

	return &user, nil
}

func (r *mongoUserRepo) ListAll(ctx context.Context) ([]*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*domain.User
	for cursor.Next(ctx) {
		var u domain.User
		if err := cursor.Decode(&u); err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	return users, nil
}
