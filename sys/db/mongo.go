package db

import (
	"context"
	"github.com/lircstar/nemo/sys/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB encapsulates the MongoDB client and database.
type MongoDB struct {
	client *mongo.Client
	db     *mongo.Database
}

// NewMongoDB creates a new MongoDB connection.
func NewMongoDB(uri, dbName string) (*MongoDB, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	// Test the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	return &MongoDB{
		client: client,
		db:     client.Database(dbName),
	}, nil
}

// GetDatabase returns the MongoDB database.
func (m *MongoDB) GetDatabase() *mongo.Database {
	return m.db
}

// GetClient returns the MongoDB client.
func (m *MongoDB) GetClient() *mongo.Client {
	return m.client
}

// GetCollection returns a collection from the MongoDB database.
func (m *MongoDB) GetCollection(name string) *mongo.Collection {
	return m.db.Collection(name)
}

// Close closes the MongoDB connection.
func (m *MongoDB) Close() error {
	return m.client.Disconnect(context.TODO())
}

// DecodeCursor decodes a MongoDB cursor into a slice of bson.M.
func (m *MongoDB) DecodeCursor(ctx context.Context, cursor *mongo.Cursor) ([]bson.M, error) {
	defer cursor.Close(ctx)
	var results []bson.M
	for cursor.Next(ctx) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

// DecodeSingleResult decodes a MongoDB single result into a bson.M.
func (m *MongoDB) DecodeSingleResult(result *mongo.SingleResult) (bson.M, error) {
	if result.Err() != nil {
		return nil, result.Err()
	}
	data := bson.M{}
	err := result.Decode(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
