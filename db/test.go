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

func CreateMockPlayerAsset[T any](data T) PlayerAsset[T] {
	return PlayerAsset[T]{
		UserID:    MockID,
		Name:      "test_image",
		AssetType: ASSET_PLAYER_UP,
		X:         0,
		Y:         0,
		Width:     16,
		Height:    16,
		Data:      data,
	}
}

func CreatePlayerAssetResponseData[T any](p PlayerAsset[T]) bson.D {
	return bson.D{
		{Key: "_id", Value: p.ID},
		{Key: "user_id", Value: p.UserID},
		{Key: "name", Value: p.Name},
		{Key: "asset_type", Value: p.AssetType},
		{Key: "x", Value: p.X},
		{Key: "y", Value: p.Y},
		{Key: "width", Value: p.Width},
		{Key: "height", Value: p.Height},
		{Key: "data", Value: p.Data},
	}
}
