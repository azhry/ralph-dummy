package database

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"wedding-invitation-backend/internal/config"
)

type MongoDBTestSuite struct {
	suite.Suite
	db  *MongoDB
	cfg *config.DatabaseConfig
}

func (suite *MongoDBTestSuite) SetupSuite() {
	suite.cfg = &config.DatabaseConfig{
		URI:      "mongodb://localhost:27017",
		Database: "wedding_test",
		Timeout:  10,
	}
}

func (suite *MongoDBTestSuite) SetupTest() {
	// Skip if MongoDB is not available
	if testing.Short() {
		suite.T().Skip("Skipping MongoDB tests in short mode")
	}

	db, err := NewMongoDB(suite.cfg)
	require.NoError(suite.T(), err)
	suite.db = db
}

func (suite *MongoDBTestSuite) TearDownTest() {
	if suite.db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := suite.db.Database.Drop(ctx)
		suite.NoError(err)

		err = suite.db.Close(ctx)
		suite.NoError(err)
	}
}

func (suite *MongoDBTestSuite) TestNewMongoDB_Success() {
	db, err := NewMongoDB(suite.cfg)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), db)
	assert.NotNil(suite.T(), db.Client)
	assert.NotNil(suite.T(), db.Database)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.Close(ctx)
	assert.NoError(suite.T(), err)
}

func (suite *MongoDBTestSuite) TestNewMongoDB_InvalidURI() {
	invalidCfg := &config.DatabaseConfig{
		URI:      "mongodb://invalid-connection-string",
		Database: "test",
		Timeout:  5,
	}

	db, err := NewMongoDB(invalidCfg)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), db)
	// The error message might vary, just check that there's an error
	assert.NotEmpty(suite.T(), err.Error())
}

func (suite *MongoDBTestSuite) TestCollection() {
	collection := suite.db.Collection("test_collection")

	assert.NotNil(suite.T(), collection)
	assert.Equal(suite.T(), "test_collection", collection.Name())
}

func (suite *MongoDBTestSuite) TestEnsureIndexes() {
	ctx := context.Background()

	err := suite.db.EnsureIndexes(ctx)
	assert.NoError(suite.T(), err)

	// Verify indexes were created
	usersCollection := suite.db.Collection("users")
	usersIndexes, err := usersCollection.Indexes().ListSpecifications(ctx)
	require.NoError(suite.T(), err)

	hasEmailIndex := false
	for _, index := range usersIndexes {
		if index.Name == "email_1" {
			hasEmailIndex = true
			break
		}
	}
	assert.True(suite.T(), hasEmailIndex, "Email index should be created")

	// Verify wedding indexes
	weddingsCollection := suite.db.Collection("weddings")
	weddingsIndexes, err := weddingsCollection.Indexes().ListSpecifications(ctx)
	require.NoError(suite.T(), err)

	hasSlugIndex := false
	hasUserIndex := false
	for _, index := range weddingsIndexes {
		if index.Name == "slug_1" {
			hasSlugIndex = true
		}
		if index.Name == "user_id_1_created_at_-1" {
			hasUserIndex = true
		}
	}
	assert.True(suite.T(), hasSlugIndex, "Slug index should be created")
	assert.True(suite.T(), hasUserIndex, "User ID index should be created")
}

func (suite *MongoDBTestSuite) TestClose() {
	ctx := context.Background()

	err := suite.db.Close(ctx)
	assert.NoError(suite.T(), err)

	// Verify connection is closed by attempting a ping
	err = suite.db.Client.Ping(ctx, nil)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "client is disconnected")
}

func (suite *MongoDBTestSuite) TestInsertAndFind() {
	ctx := context.Background()

	collection := suite.db.Collection("test_documents")

	doc := map[string]interface{}{
		"name":        "Test Document",
		"description": "This is a test document",
		"created_at":  time.Now(),
	}

	result, err := collection.InsertOne(ctx, doc)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result.InsertedID)

	// Find the document
	var found bson.M
	err = collection.FindOne(ctx, map[string]interface{}{"_id": result.InsertedID}).Decode(&found)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Test Document", found["name"])
	assert.Equal(suite.T(), "This is a test document", found["description"])
}

func TestMongoDBTestSuite(t *testing.T) {
	suite.Run(t, new(MongoDBTestSuite))
}

// Integration test with real MongoDB
func TestMongoDB_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &config.DatabaseConfig{
		URI:      "mongodb://localhost:27017",
		Database: "wedding_integration_test",
		Timeout:  10,
	}

	db, err := NewMongoDB(cfg)
	require.NoError(t, err)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		db.Database.Drop(ctx)
		db.Close(ctx)
	}()

	// Test basic operations
	collection := db.Collection("integration_test")
	doc := map[string]interface{}{
		"test":  "integration",
		"value": 42,
	}

	result, err := collection.InsertOne(context.Background(), doc)
	require.NoError(t, err)
	assert.NotNil(t, result.InsertedID)

	var found bson.M
	err = collection.FindOne(context.Background(), map[string]interface{}{"_id": result.InsertedID}).Decode(&found)
	require.NoError(t, err)
	assert.Equal(t, "integration", found["test"])
	assert.Equal(t, int32(42), found["value"])
}
