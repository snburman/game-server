package db

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

const (
	MockID = "67bfa82f165e6e4169699147"
)

func NewMockMongoDriver(client *mongo.Client) *MongoDriver {
	return &MongoDriver{
		Client: client,
	}
}

var SuccessResponse = mtest.CreateSuccessResponse(
	bson.D{
		{Key: "ok", Value: 1},
		{Key: "acknowledged", Value: true},
		{Key: "n", Value: 1},
	}...,
)

func CreateCursorEnd(dbTable string) bson.D {
	return mtest.CreateCursorResponse(
		0,
		dbTable,
		mtest.NextBatch,
	)
}
