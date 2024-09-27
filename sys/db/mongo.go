package db

import (
	"context"
	"github.com/lircstar/nemo/sys/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConnection struct {
	client *mongo.Client
	db     *mongo.Database
}

func NewMongoConnection(uri string, db string) (*MongoConnection, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	// 测试连接
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	return &MongoConnection{
		client: client,
		db:     client.Database(db),
	}, nil
}
func (m *MongoConnection) GetClient() *mongo.Client {
	return m.client
}

func (m *MongoConnection) GetDatabase() *mongo.Database {
	return m.db
}

func (m *MongoConnection) Close() error {
	return m.client.Disconnect(context.TODO())
}

type Collection struct {
	collection *mongo.Collection
}

// NewCollection 创建一个新的 Collection 实例
func (m *MongoConnection) NewCollection(name string) *Collection {
	return &Collection{
		collection: m.db.Collection(name),
	}
}

// InsertOne 插入一条文档
func (c *Collection) InsertOne(ctx context.Context, document any) (*mongo.InsertOneResult, error) {
	return c.collection.InsertOne(ctx, document)
}

// Find 查询文档
func (c *Collection) Find(ctx context.Context, filter any) ([]bson.M, error) {
	cur, err := c.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var results []bson.M
	for cur.Next(ctx) {
		var result bson.M
		if err := cur.Decode(&result); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// UpdateOne update a document
func (c *Collection) UpdateOne(ctx context.Context, filter, update any) (*mongo.UpdateResult, error) {
	return c.collection.UpdateOne(ctx, filter, update)
}

// DeleteOne delete a document
func (c *Collection) DeleteOne(ctx context.Context, filter any) (*mongo.DeleteResult, error) {
	return c.collection.DeleteOne(ctx, filter)
}

// ExecutePipeline execute an aggregation pipeline
func (c *Collection) ExecutePipeline(ctx context.Context, pipeline any) ([]bson.M, error) {
	cur, err := c.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var results []bson.M
	for cur.Next(ctx) {
		var result bson.M
		if err := cur.Decode(&result); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
