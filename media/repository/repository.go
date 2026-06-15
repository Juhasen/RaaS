package repository

import (
	"context"
	"errors"

	"github.com/Juhasen/RaaS/media/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// MediaRepository defines the database operations for media metadata.
type MediaRepository interface {
	Save(ctx context.Context, media *models.Media) error
	FindByID(ctx context.Context, id string) (*models.Media, error)
	FindByListingID(ctx context.Context, listingID string) ([]*models.Media, error)
	DeleteByListingID(ctx context.Context, listingID string) error
}

// MongoMediaRepository is a MongoDB implementation of MediaRepository.
type MongoMediaRepository struct {
	collection *mongo.Collection
}

// NewMongoMediaRepository creates a new MongoMediaRepository instance.
func NewMongoMediaRepository(client *mongo.Client, dbName string) *MongoMediaRepository {
	return &MongoMediaRepository{
		collection: client.Database(dbName).Collection("media"),
	}
}

// Save persists a media record in MongoDB.
func (r *MongoMediaRepository) Save(ctx context.Context, media *models.Media) error {
	_, err := r.collection.InsertOne(ctx, media)
	return err
}

// FindByID retrieves a media record by its ID.
func (r *MongoMediaRepository) FindByID(ctx context.Context, id string) (*models.Media, error) {
	var media models.Media
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&media)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &media, nil
}

// FindByListingID retrieves all media records associated with a listing ID.
func (r *MongoMediaRepository) FindByListingID(ctx context.Context, listingID string) ([]*models.Media, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"listing_id": listingID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []*models.Media
	for cursor.Next(ctx) {
		var media models.Media
		if err := cursor.Decode(&media); err != nil {
			return nil, err
		}
		list = append(list, &media)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

// DeleteByListingID removes all media records for a given listing ID.
func (r *MongoMediaRepository) DeleteByListingID(ctx context.Context, listingID string) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"listing_id": listingID})
	return err
}
