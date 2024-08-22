package db

import (
	"context"
	"github.com/lircstar/nemo/sys/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConnection struct {
	client *mongo.Client
	db     *mongo.Database
}

func ConnectMongoDB(uri string, db string) (*MongoConnection, error) {
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

func (m *MongoConnection) GetCollection(name string) *mongo.Collection {
	return m.db.Collection(name)
}

func (m *MongoConnection) Close() {
	err := m.client.Disconnect(context.TODO())
	if err != nil {
		log.Error(err)
	}
}
