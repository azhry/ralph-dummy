package database

import (
	"context"
	"fmt"
	"time"

	"wedding-invitation-backend/internal/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func NewMongoDB(cfg *config.DatabaseConfig) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Timeout)*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.URI)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return &MongoDB{
		Client:   client,
		Database: client.Database(cfg.Database),
	}, nil
}

func (m *MongoDB) Close(ctx context.Context) error {
	return m.Client.Disconnect(ctx)
}

func (m *MongoDB) Collection(name string) *mongo.Collection {
	return m.Database.Collection(name)
}

func (m *MongoDB) EnsureIndexes(ctx context.Context) error {
	users := m.Collection("users")
	if _, err := users.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return fmt.Errorf("failed to create users email index: %w", err)
	}

	weddings := m.Collection("weddings")
	if _, err := weddings.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "slug", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return fmt.Errorf("failed to create weddings slug index: %w", err)
	}

	if _, err := weddings.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}},
	}); err != nil {
		return fmt.Errorf("failed to create weddings user_id index: %w", err)
	}

	// RSVP indexes
	rsvps := m.Collection("rsvps")
	if _, err := rsvps.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "wedding_id", Value: 1}, {Key: "submitted_at", Value: -1}},
	}); err != nil {
		return fmt.Errorf("failed to create rsvps wedding_id index: %w", err)
	}

	if _, err := rsvps.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "wedding_id", Value: 1}, {Key: "email", Value: 1}},
	}); err != nil {
		return fmt.Errorf("failed to create rsvps email index: %w", err)
	}

	// Guest indexes
	guests := m.Collection("guests")
	if _, err := guests.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "wedding_id", Value: 1}, {Key: "created_at", Value: -1}},
	}); err != nil {
		return fmt.Errorf("failed to create guests wedding_id index: %w", err)
	}

	if _, err := guests.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "wedding_id", Value: 1}, {Key: "email", Value: 1}},
	}); err != nil {
		return fmt.Errorf("failed to create guests email index: %w", err)
	}

	// Analytics indexes
	pageViews := m.Collection("page_views")
	if _, err := pageViews.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "wedding_id", Value: 1}, {Key: "timestamp", Value: -1}},
	}); err != nil {
		return fmt.Errorf("failed to create page_views wedding_id index: %w", err)
	}

	if _, err := pageViews.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "session_id", Value: 1}},
	}); err != nil {
		return fmt.Errorf("failed to create page_views session_id index: %w", err)
	}

	if _, err := pageViews.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "wedding_id", Value: 1}, {Key: "page", Value: 1}},
	}); err != nil {
		return fmt.Errorf("failed to create page_views page index: %w", err)
	}

	if _, err := pageViews.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "timestamp", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(7776000), // 90 days TTL
	}); err != nil {
		return fmt.Errorf("failed to create page_views TTL index: %w", err)
	}

	// RSVP analytics indexes
	rsvpAnalytics := m.Collection("rsvp_analytics")
	if _, err := rsvpAnalytics.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "wedding_id", Value: 1}, {Key: "timestamp", Value: -1}},
	}); err != nil {
		return fmt.Errorf("failed to create rsvp_analytics wedding_id index: %w", err)
	}

	if _, err := rsvpAnalytics.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "session_id", Value: 1}},
	}); err != nil {
		return fmt.Errorf("failed to create rsvp_analytics session_id index: %w", err)
	}

	if _, err := rsvpAnalytics.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "timestamp", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(7776000), // 90 days TTL
	}); err != nil {
		return fmt.Errorf("failed to create rsvp_analytics TTL index: %w", err)
	}

	// Conversion events indexes
	conversions := m.Collection("conversion_events")
	if _, err := conversions.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "wedding_id", Value: 1}, {Key: "timestamp", Value: -1}},
	}); err != nil {
		return fmt.Errorf("failed to create conversion_events wedding_id index: %w", err)
	}

	if _, err := conversions.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "session_id", Value: 1}},
	}); err != nil {
		return fmt.Errorf("failed to create conversion_events session_id index: %w", err)
	}

	if _, err := conversions.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "event", Value: 1}},
	}); err != nil {
		return fmt.Errorf("failed to create conversion_events event index: %w", err)
	}

	if _, err := conversions.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "timestamp", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(7776000), // 90 days TTL
	}); err != nil {
		return fmt.Errorf("failed to create conversion_events TTL index: %w", err)
	}

	// Wedding analytics indexes
	weddingAnalytics := m.Collection("wedding_analytics")
	if _, err := weddingAnalytics.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return fmt.Errorf("failed to create wedding_analytics _id index: %w", err)
	}

	if _, err := weddingAnalytics.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "last_updated", Value: -1}},
	}); err != nil {
		return fmt.Errorf("failed to create wedding_analytics last_updated index: %w", err)
	}

	// System analytics indexes
	systemAnalytics := m.Collection("system_analytics")
	if _, err := systemAnalytics.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return fmt.Errorf("failed to create system_analytics _id index: %w", err)
	}

	return nil
}
