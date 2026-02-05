package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"wedding-invitation-backend/internal/config"
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
		Keys:    map[string]interface{}{"email": 1},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return fmt.Errorf("failed to create users email index: %w", err)
	}

	weddings := m.Collection("weddings")
	if _, err := weddings.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    map[string]interface{}{"slug": 1},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return fmt.Errorf("failed to create weddings slug index: %w", err)
	}

	if _, err := weddings.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: map[string]interface{}{"user_id": 1, "created_at": -1},
	}); err != nil {
		return fmt.Errorf("failed to create weddings user_id index: %w", err)
	}

	return nil
}
