package db

import (
	"testing"

	"github.com/snburman/game-server/errors"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

var assetSource = "game.player_images"

func createMockPlayerAsset[T any](data T) PlayerAsset[T] {
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

func createPlayerAssetResponseData[T any](p PlayerAsset[T]) bson.D {
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

func TestCreatePlayerAsset(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mockPlayerAsset := createMockPlayerAsset("test_data")
	mt.Run("success", func(mt *mtest.T) {
		// arrange for success
		mt.AddMockResponses(
			// first find operation returns no results
			mtest.CreateCursorResponse(
				0,
				assetSource,
				mtest.FirstBatch,
			),
			// insert operation is successful
			mtest.CreateSuccessResponse(
				bson.D{
					{Key: "ok", Value: 1},
					{Key: "acknowledged", Value: true},
					{Key: "n", Value: 1},
				}...,
			),
		)
		// act
		driver := NewMockMongoDriver(mt.Client)
		res, err := CreatePlayerAsset(driver, mockPlayerAsset)

		// assert
		assert.Nil(t, err, "expected nil but got error")
		assert.NotEqual(t, res, primitive.NilObjectID, "expected non-nil object ID")
	})
	mt.Run("failure-image-exists", func(mt *mtest.T) {
		// arrange for failure
		mt.AddMockResponses(
			// find operation returns existing document
			// and fails CreatePlayerAsset
			mtest.CreateCursorResponse(
				1,
				assetSource,
				mtest.FirstBatch,
				createPlayerAssetResponseData(mockPlayerAsset),
			),
		)
		// act
		driver := NewMockMongoDriver(mt.Client)
		res, err := CreatePlayerAsset(driver, mockPlayerAsset)

		// assert
		assert.NotNil(t, err, "expected error but got nil")
		assert.Equal(t, err, errors.ErrImageExists, "expected ErrImageExists")
		assert.Equal(t, res, primitive.NilObjectID, "expected nil object ID")
	})
}

func TestGetPlayerAssetsByUserID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mockPlayerAsset := createMockPlayerAsset([]byte("[]"))
	mt.Run("success", func(mt *mtest.T) {
		// arrange for success
		mt.AddMockResponses(
			// find operation is successful
			mtest.CreateCursorResponse(
				1,
				assetSource,
				mtest.FirstBatch,
				createPlayerAssetResponseData(mockPlayerAsset),
			),
			mtest.CreateCursorResponse(0, assetSource, mtest.NextBatch),
		)
		// act
		driver := NewMockMongoDriver(mt.Client)
		res, err := GetPlayerAssetsByUserID(driver, mockPlayerAsset.UserID)

		// assert nil error
		assert.Nil(t, err, "expected nil but got error")
		// assert data is of type []PlayerAsset[PixelData]
		assert.IsType(
			t,
			[]PlayerAsset[PixelData]{},
			res,
			"expected []PlayerAsset[PixelData] but got different type",
		)
		// assert fields of returned data
		assert.Equal(t, res[0].UserID, mockPlayerAsset.UserID, "expected same user ID")
		assert.Equal(t, res[0].Name, mockPlayerAsset.Name, "expected same name")
		assert.Equal(t, res[0].AssetType, mockPlayerAsset.AssetType, "expected same asset type")
		assert.Equal(t, res[0].X, mockPlayerAsset.X, "expected same X")
		assert.Equal(t, res[0].Y, mockPlayerAsset.Y, "expected same Y")
		assert.Equal(t, res[0].Width, mockPlayerAsset.Width, "expected same width")
		assert.Equal(t, res[0].Height, mockPlayerAsset.Height, "expected same height")
	})

	mt.Run("failure-wrong-format", func(mt *mtest.T) {
		mockPlayerAsset := createMockPlayerAsset("wrong_data")
		// arrange for failure
		mt.AddMockResponses(
			// find operation fails
			mtest.CreateCursorResponse(
				1,
				assetSource,
				mtest.FirstBatch,
				createPlayerAssetResponseData(mockPlayerAsset),
			),
			mtest.CreateCursorResponse(0, assetSource, mtest.NextBatch),
		)
		driver := NewMockMongoDriver(mt.Client)
		// act
		_, err := GetPlayerAssetsByUserID(driver, mockPlayerAsset.UserID)
		// assert error
		assert.NotNil(t, err, "expected error but got nil")
		assert.Equal(t, err, errors.ErrImageWrongFormat, "expected ErrImageWrongFormat")
	})
}

func TestGetPlayerCharactersByUserIDs(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mockPlayerAsset := createMockPlayerAsset([]byte("[]"))
	mt.Run("success", func(mt *mtest.T) {
		// arrange for success
		mt.AddMockResponses(
			// find operation is successful
			mtest.CreateCursorResponse(
				1,
				assetSource,
				mtest.FirstBatch,
				createPlayerAssetResponseData(mockPlayerAsset),
			),
			CreateCursorEnd(assetSource),
		)
		// act
		driver := NewMockMongoDriver(mt.Client)
		_, err := GetPlayerCharactersByUserIDs(driver, []string{mockPlayerAsset.UserID})

		// assert
		assert.Nil(t, err, "expected nil but got error")
	})

	mt.Run("failure-wrong-format", func(mt *mtest.T) {
		mockPlayerAsset := createMockPlayerAsset("wrong_data")
		// arrange for failure
		mt.AddMockResponses(
			// find operation fails
			mtest.CreateCursorResponse(
				1,
				assetSource,
				mtest.FirstBatch,
				createPlayerAssetResponseData(mockPlayerAsset),
			),
			CreateCursorEnd(assetSource),
		)
		driver := NewMockMongoDriver(mt.Client)
		// act
		_, err := GetPlayerCharactersByUserIDs(driver, []string{mockPlayerAsset.UserID})
		// assert error
		assert.NotNil(t, err, "expected error but got nil")
		assert.Equal(t, err, errors.ErrImageWrongFormat, "expected ErrImageWrongFormat")
	})
}

func TestAppendMapPlayerCharacter(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mockPlayerAsset := createMockPlayerAsset([]byte("[]"))
	mt.Run("success", func(mt *mtest.T) {
		// arrange for success
		mt.AddMockResponses(
			mtest.CreateCursorResponse(
				1,
				assetSource,
				mtest.FirstBatch,
				createPlayerAssetResponseData(mockPlayerAsset),
			),
			CreateCursorEnd(assetSource),
		)
		_map := Map[[]PlayerAsset[PixelData]]{
			Data: []PlayerAsset[PixelData]{},
		}
		// act
		driver := NewMockMongoDriver(mt.Client)
		_, err := AppendMapPlayerCharacter(driver, mockPlayerAsset.UserID, _map)

		// assert
		assert.Nil(t, err, "expected nil but got error")
	})

	mt.Run("failure", func(mt *mtest.T) {
		mockPlayerAsset := createMockPlayerAsset("wrong_data")
		// arrange for failure
		mt.AddMockResponses(
			// find operation fails
			mtest.CreateCursorResponse(
				0,
				assetSource,
				mtest.FirstBatch,
				createPlayerAssetResponseData(mockPlayerAsset),
			),
			CreateCursorEnd(assetSource),
		)
		_map := Map[[]PlayerAsset[PixelData]]{
			Data: []PlayerAsset[PixelData]{},
		}
		// act
		driver := NewMockMongoDriver(mt.Client)
		_, err := AppendMapPlayerCharacter(driver, mockPlayerAsset.UserID, _map)

		// assert
		assert.NotNil(t, err, "expected error but got nil")
		assert.Equal(t, err, errors.ErrImageWrongFormat, "expected ErrImageWrongFormat")
	})
}

func TestGetPlayerAssetByNameUserID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mockPlayerAsset := createMockPlayerAsset([]byte("[]"))
	mt.Run("success", func(mt *mtest.T) {
		// arrange for success
		mt.AddMockResponses(
			mtest.CreateCursorResponse(
				1,
				assetSource,
				mtest.FirstBatch,
				createPlayerAssetResponseData(mockPlayerAsset),
			),
		)
		// act
		driver := NewMockMongoDriver(mt.Client)
		_, err := GetPlayerAssetByNameUserID(driver, mockPlayerAsset.Name, mockPlayerAsset.UserID)
		// assert
		assert.Nil(t, err, "expected nil but got error")
	})

	mt.Run("failure-not-found", func(mt *mtest.T) {
		// arrange for failure
		mt.AddMockResponses(
			// find operation returns no results
			mtest.CreateCursorResponse(
				0,
				assetSource,
				mtest.FirstBatch,
			),
		)
		// act
		driver := NewMockMongoDriver(mt.Client)
		_, err := GetPlayerAssetByNameUserID(driver, "", "")

		// assert
		assert.NotNil(t, err, "expected error but got nil")
		assert.Equal(t, err, errors.ErrImageNotFound, "expected ErrImageNotFound")
	})

	mt.Run("failure-wrong-format", func(mt *mtest.T) {
		mockPlayerAsset := createMockPlayerAsset("wrong_data")
		// arrange for failure
		mt.AddMockResponses(
			// find operation fails
			mtest.CreateCursorResponse(
				1,
				assetSource,
				mtest.FirstBatch,
				createPlayerAssetResponseData(mockPlayerAsset),
			),
		)
		// act
		driver := NewMockMongoDriver(mt.Client)
		_, err := GetPlayerAssetByNameUserID(driver, mockPlayerAsset.Name, mockPlayerAsset.UserID)

		// assert
		assert.NotNil(t, err, "expected error but got nil")
		assert.Equal(t, err, errors.ErrImageWrongFormat, "expected ErrImageWrongFormat")
	})
}

func TestGetDefaultPlayerCharacter(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mockPlayerAsset := createMockPlayerAsset([]byte("[]"))
	mt.Run("success", func(mt *mtest.T) {
		mt.AddMockResponses(
			// find operation is successful
			mtest.CreateCursorResponse(
				1,
				assetSource,
				mtest.FirstBatch,
				createPlayerAssetResponseData(mockPlayerAsset),
			),
		)
		driver := NewMockMongoDriver(mt.Client)
		_, err := GetDefaultPlayerCharacter(driver)
		if err != nil {
			t.Fatalf("GetDefaultPlayerCharacter failed: %v", err)
		}
	})
}

func TestUpdatePlayerAsset(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mockPlayerAsset := createMockPlayerAsset("test_data")
	mt.Run("success", func(mt *mtest.T) {
		// arrange for success
		mt.AddMockResponses(
			mtest.CreateCursorResponse(
				1,
				assetSource,
				mtest.FirstBatch,
				createPlayerAssetResponseData(mockPlayerAsset),
			),
			// update operation is successful
			SuccessResponse,
		)
		// act
		driver := NewMockMongoDriver(mt.Client)
		err := UpdatePlayerAsset(driver, mockPlayerAsset)

		// assert
		assert.Nil(t, err, "expected nil but got error")
	})
}

func TestDeletePlayerAsset(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mockPlayerAsset := createMockPlayerAsset("test_data")
	mt.Run("success", func(mt *mtest.T) {
		mt.AddMockResponses(
			// find operation is successful
			mtest.CreateCursorResponse(
				1,
				assetSource,
				mtest.FirstBatch,
				createPlayerAssetResponseData(mockPlayerAsset),
			),
			// delete operation is successful
			SuccessResponse,
		)
		driver := NewMockMongoDriver(mt.Client)
		_, err := DeletePlayerAsset(driver, mockPlayerAsset.ID.Hex())
		if err != nil {
			t.Fatalf("DeletePlayerAsset failed: %v", err)
		}
	})

	mt.Run("failure", func(mt *mtest.T) {
		driver := NewMockMongoDriver(mt.Client)
		_, err := DeletePlayerAsset(driver, "wrong_id_format")
		// assert
		assert.NotNil(t, err, "expected error but got nil")
	})
}
