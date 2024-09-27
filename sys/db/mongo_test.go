package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func TestMongoDBOperations(t *testing.T) {

	uri := "mongodb://localhost:27017"
	client, err := NewMongoConnection(uri, "test")
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err = client.Close(); err != nil {
			t.Fatalf("Failed to disconnect MongoDB: %v", err)
		}
	}()

	users := client.NewCollection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Insert document
	user := bson.M{
		"name":  "Alice",
		"age":   30,
		"email": "alice@example.com",
	}
	insertResult, err := users.InsertOne(ctx, user)
	if err != nil {
		t.Fatalf("InsertOne error: %v", err)
	}
	fmt.Printf("Inserted document with _id: %v\n", insertResult.InsertedID)

	// Find document
	filter := bson.M{"name": "Alice"}
	foundUsers, err := users.Find(ctx, filter)
	if err != nil {
		t.Fatalf("Find error: %v", err)
	}
	fmt.Printf("Found users: %v\n", foundUsers)

	// Update document
	update := bson.M{
		"$set": bson.M{
			"age": 31,
		},
	}
	updateResult, err := users.UpdateOne(ctx, filter, update)
	if err != nil {
		t.Fatalf("UpdateOne error: %v", err)
	}
	fmt.Printf("Matched %d documents and updated %d documents.\n", updateResult.MatchedCount, updateResult.ModifiedCount)

	// Delete document
	deleteResult, err := users.DeleteOne(ctx, filter)
	if err != nil {
		t.Fatalf("DeleteOne error: %v", err)
	}
	fmt.Printf("Deleted %d documents.\n", deleteResult.DeletedCount)

	// Execute aggregation pipeline
	pipeline := []bson.M{
		{"$match": bson.M{"age": bson.M{"$gte": 18}}},
		{"$group": bson.M{"_id": "$age", "count": bson.M{"$sum": 1}}},
	}
	aggResults, err := users.ExecutePipeline(ctx, pipeline)
	if err != nil {
		t.Fatalf("ExecutePipeline error: %v", err)
	}
	fmt.Printf("Aggregation results: %v\n", aggResults)
}
