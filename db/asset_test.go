package db

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

var mockPlayerAsset = PlayerAsset[string]{
	UserID:    "67bfa82f165e6e4169699147",
	Name:      "test_image",
	AssetType: ASSET_TILE,
	X:         0,
	Y:         0,
	Width:     16,
	Height:    16,
	Data:      "data",
}

func NewMockMongoDriver(client *mongo.Client) *MongoDriver {
	return &MongoDriver{
		Client: client,
	}
}

func TestCreatePlayerAsset(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		mt.AddMockResponses(
			// first find operation returns no results
			mtest.CreateCursorResponse(0, "game.player_images", mtest.FirstBatch),
			// insert operation is successful
			mtest.CreateSuccessResponse(
				bson.D{
					{Key: "ok", Value: 1},
					{Key: "acknowledged", Value: true},
					{Key: "n", Value: 1},
				}...,
			),
			// following find operation returns the inserted document
			// and fails CreatePlayerAsset
			mtest.CreateCursorResponse(1, "game.player_images", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: mockPlayerAsset.ID},
				{Key: "user_id", Value: mockPlayerAsset.UserID},
				{Key: "name", Value: mockPlayerAsset.Name},
				{Key: "asset_type", Value: mockPlayerAsset.AssetType},
				{Key: "x", Value: mockPlayerAsset.X},
				{Key: "y", Value: mockPlayerAsset.Y},
				{Key: "width", Value: mockPlayerAsset.Width},
				{Key: "height", Value: mockPlayerAsset.Height},
				{Key: "data", Value: mockPlayerAsset.Data},
			}),
		)
		driver := NewMockMongoDriver(mt.Client)
		_, err := CreatePlayerAsset(driver, mockPlayerAsset)
		if err != nil {
			t.Fatalf("CreatePlayerAsset failed: %v", err)
		}

		_, err = CreatePlayerAsset(driver, mockPlayerAsset)
		if err == nil {
			t.Fatal("expected CreatePlayerAsset to fail")
		}
	})
}
