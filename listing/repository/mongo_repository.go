package repository

import (
	"context"
	"time"

	"listing/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoRepository handles direct database operations for the listing service.
type MongoRepository struct {
	client *mongo.Client
	dbName string
}

// NewMongoRepository initializes and returns a MongoRepository instance.
func NewMongoRepository(client *mongo.Client, dbName string) *MongoRepository {
	return &MongoRepository{
		client: client,
		dbName: dbName,
	}
}

func (r *MongoRepository) GetListingsCollection() *mongo.Collection {
	return r.client.Database(r.dbName).Collection("listings")
}

func (r *MongoRepository) GetBlocksCollection() *mongo.Collection {
	return r.client.Database(r.dbName).Collection("availability_blocks")
}

func (r *MongoRepository) GetTombstonesCollection() *mongo.Collection {
	return r.client.Database(r.dbName).Collection("cancellation_tombstones")
}

// CreateListing inserts a new Listing into database.
func (r *MongoRepository) CreateListing(ctx context.Context, listing *models.Listing) error {
	coll := r.GetListingsCollection()
	_, err := coll.InsertOne(ctx, listing)
	return err
}

// GetListingByID retrieves a Listing by its ID.
func (r *MongoRepository) GetListingByID(ctx context.Context, id primitive.ObjectID) (*models.Listing, error) {
	coll := r.GetListingsCollection()
	var l models.Listing
	err := coll.FindOne(ctx, bson.M{"_id": id}).Decode(&l)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

// UpdateListing replaces an existing Listing document.
func (r *MongoRepository) UpdateListing(ctx context.Context, listing *models.Listing) error {
	coll := r.GetListingsCollection()
	_, err := coll.ReplaceOne(ctx, bson.M{"_id": listing.ID}, listing)
	return err
}

// DeleteListing deletes a Listing by its ID.
func (r *MongoRepository) DeleteListing(ctx context.Context, id primitive.ObjectID) error {
	coll := r.GetListingsCollection()
	res, err := coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

// ListListings retrieves all Listings that match the provided BSON filter query.
func (r *MongoRepository) ListListings(ctx context.Context, filter bson.M) ([]models.Listing, error) {
	coll := r.GetListingsCollection()
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []models.Listing
	if err := cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	// Return empty slice instead of nil for clean JSON arrays
	if list == nil {
		list = []models.Listing{}
	}
	return list, nil
}

// UpsertAvailabilityBlock upserts an AvailabilityBlock document keyed by booking_id+date.
func (r *MongoRepository) UpsertAvailabilityBlock(ctx context.Context, block *models.AvailabilityBlock) error {
	coll := r.GetBlocksCollection()
	filter := bson.M{
		"listing_id": block.ListingID,
		"date":       block.Date,
		"booking_id": block.BookingID,
	}
	_, err := coll.UpdateOne(ctx, filter, bson.M{"$setOnInsert": block}, options.Update().SetUpsert(true))
	return err
}

// DeleteAvailabilityBlocksByBooking deletes availability blocks for a given booking ID.
func (r *MongoRepository) DeleteAvailabilityBlocksByBooking(ctx context.Context, bookingID string) (int64, error) {
	coll := r.GetBlocksCollection()
	res, err := coll.DeleteMany(ctx, bson.M{"booking_id": bookingID})
	if err != nil {
		return 0, err
	}
	return res.DeletedCount, nil
}

// CountAvailabilityBlocks counts blocking blocks for a listing within a date range [start, end).
func (r *MongoRepository) CountAvailabilityBlocks(ctx context.Context, listingID primitive.ObjectID, start, end time.Time) (int64, error) {
	coll := r.GetBlocksCollection()
	filter := bson.M{
		"listing_id": listingID,
		"date": bson.M{
			"$gte": start,
			"$lt":  end,
		},
	}
	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetBlockedListingIDs returns listing IDs that have availability blocks in the range [start, end).
func (r *MongoRepository) GetBlockedListingIDs(ctx context.Context, start, end time.Time) ([]primitive.ObjectID, error) {
	coll := r.GetBlocksCollection()
	filter := bson.M{
		"date": bson.M{
			"$gte": start,
			"$lt":  end,
		},
	}
	values, err := coll.Distinct(ctx, "listing_id", filter)
	if err != nil {
		return nil, err
	}
	var blockedIDs []primitive.ObjectID
	for _, val := range values {
		if oid, ok := val.(primitive.ObjectID); ok {
			blockedIDs = append(blockedIDs, oid)
		}
	}
	if blockedIDs == nil {
		blockedIDs = []primitive.ObjectID{}
	}
	return blockedIDs, nil
}

// CreateTombstone inserts a cancellation tombstone to prevent late bookings.
func (r *MongoRepository) CreateTombstone(ctx context.Context, tombstone *models.CancellationTombstone) error {
	coll := r.GetTombstonesCollection()
	filter := bson.M{"booking_id": tombstone.BookingID}
	_, err := coll.UpdateOne(ctx, filter, bson.M{"$setOnInsert": tombstone}, options.Update().SetUpsert(true))
	return err
}

// DeleteTombstoneByBooking removes a cancellation tombstone if it exists.
func (r *MongoRepository) DeleteTombstoneByBooking(ctx context.Context, bookingID string) (int64, error) {
	coll := r.GetTombstonesCollection()
	res, err := coll.DeleteOne(ctx, bson.M{"booking_id": bookingID})
	if err != nil {
		return 0, err
	}
	return res.DeletedCount, nil
}
