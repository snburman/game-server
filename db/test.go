package db

import "go.mongodb.org/mongo-driver/mongo"

func NewMockMongoDriver(client *mongo.Client) *MongoDriver {
	return &MongoDriver{
		Client: client,
	}
}
