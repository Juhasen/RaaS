package main

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AvailabilityBlock struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	ListingID primitive.ObjectID `bson:"listing_id"`
	Date      time.Time          `bson:"date"`
	BookingID string             `bson:"booking_id"`
	CreatedAt time.Time          `bson:"created_at"`
}

type CancellationTombstone struct {
	BookingID string    `bson:"booking_id"`
	CreatedAt time.Time `bson:"created_at"`
}

// processBookingConfirmed inserts per-day availability blocks for the booking.
func processBookingConfirmed(listingIDHex, bookingID, startStr, endStr string) error {
	if MongoClient == nil {
		return fmt.Errorf("mongo client nil")
	}
	listingID, err := primitive.ObjectIDFromHex(listingIDHex)
	if err != nil {
		return err
	}
	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return err
	}
	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return err
	}

	blocksColl := MongoClient.Database("raas").Collection("availability_blocks")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// For each day in [start, end) insert an upserted block keyed by booking_id+date
	for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
		filter := bson.M{"listing_id": listingID, "date": d, "booking_id": bookingID}
		doc := bson.M{"listing_id": listingID, "date": d, "booking_id": bookingID, "created_at": time.Now()}
		_, err := blocksColl.UpdateOne(ctx, filter, bson.M{"$setOnInsert": doc}, options.Update().SetUpsert(true))
		if err != nil {
			return err
		}
	}

	// If there is a tombstone for this booking, remove it because we have a confirmation later than cancellation.
	tombColl := MongoClient.Database("raas").Collection("cancellation_tombstones")
	_, _ = tombColl.DeleteOne(ctx, bson.M{"booking_id": bookingID})

	return nil
}

// processBookingCancelled removes blocks for a given booking. If none found, record a tombstone.
func processBookingCancelled(listingIDHex, bookingID string) error {
	if MongoClient == nil {
		return fmt.Errorf("mongo client nil")
	}
	// listingID may be unknown for cancellation; we still attempt to delete by booking_id
	blocksColl := MongoClient.Database("raas").Collection("availability_blocks")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := blocksColl.DeleteMany(ctx, bson.M{"booking_id": bookingID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		// record tombstone so future confirmations are ignored
		tombColl := MongoClient.Database("raas").Collection("cancellation_tombstones")
		_, err := tombColl.UpdateOne(ctx, bson.M{"booking_id": bookingID}, bson.M{"$setOnInsert": bson.M{"booking_id": bookingID, "created_at": time.Now()}}, options.Update().SetUpsert(true))
		if err != nil {
			return err
		}
	}
	return nil
}
