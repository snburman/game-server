package db

import (
	"testing"

	"github.com/snburman/game-server/errors"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

var mapSource = "game.player_maps"

func createMockMap[T any](data T) Map[T] {
	return Map[T]{
		Data:     data,
		Name:     "name",
		Primary:  true,
		UserID:   MockID,
		UserName: "userName",
	}
}

func createMapResponseData[T any](m Map[T]) bson.D {
	return bson.D{
		{Key: "_id", Value: m.ID},
		{Key: "user_id", Value: m.UserID},
		{Key: "username", Value: m.UserName},
		{Key: "name", Value: m.Name},
		{Key: "primary", Value: m.Primary},
		{Key: "entrance", Value: bson.D{
			{Key: "x", Value: m.Entrance.X},
			{Key: "y", Value: m.Entrance.Y},
		}},
		{Key: "portals", Value: m.Portals},
		{Key: "data", Value: m.Data},
	}
}

func TestCreateMap(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mockMap := createMockMap("data")

	mt.Run("success", func(mt *mtest.T) {
		// arrange for success
		mt.AddMockResponses(
			mtest.CreateCursorResponse(
				0,
				mapSource,
				mtest.FirstBatch,
			),
			mtest.CreateCursorResponse(
				0,
				mapSource,
				mtest.FirstBatch,
			),
			mtest.CreateCursorResponse(
				1,
				mapSource,
				mtest.FirstBatch,
				createMapResponseData(mockMap),
			),
			SuccessResponse,
		)
		// act
		driver := NewMockMongoDriver(mt.Client)
		id, err := CreateMap(driver, mockMap)

		// assert
		assert.Nil(t, err)
		assert.NotEqual(t, id, primitive.NilObjectID, "expected id to be non-zero")
	})

	mt.Run("failure-map-exists", func(mt *mtest.T) {
		// arrange for failure
		mt.AddMockResponses(
			mtest.CreateCursorResponse(
				1,
				mapSource,
				mtest.FirstBatch,
				createMapResponseData(mockMap),
			),
		)
		// act
		driver := NewMockMongoDriver(mt.Client)
		_, err := CreateMap(driver, mockMap)

		// assert
		assert.NotNil(t, err, "expected error to be nil")
		assert.Equal(t, err, errors.ErrMapExists)
	})

	mt.Run("failure-wrong-objectID-format", func(mt *mtest.T) {
		// arrange for failure
		mt.AddMockResponses(
			mtest.CreateCursorResponse(
				0,
				mapSource,
				mtest.FirstBatch,
			),
			mtest.CreateCursorResponse(
				1,
				mapSource,
				mtest.FirstBatch,
				createMapResponseData(mockMap),
			),
			mtest.CreateCursorResponse(
				0,
				mapSource,
				mtest.NextBatch,
			),
		)
		// act
		driver := NewMockMongoDriver(mt.Client)
		_, err := CreateMap(driver, mockMap)

		// assert
		assert.NotNil(t, err)
		assert.Equal(t, err, errors.ErrCreatingMap)
	})
}

func TestGetAllMaps(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mockMap := createMockMap([]byte(("[]")))

	mt.Run("success", func(mt *mtest.T) {
		// arrange for success
		mt.AddMockResponses(
			mtest.CreateCursorResponse(
				1,
				mapSource,
				mtest.FirstBatch,
				createMapResponseData(mockMap),
			),
			mtest.CreateCursorResponse(0, mapSource, mtest.NextBatch),
		)
		// act
		driver := NewMockMongoDriver(mt.Client)
		maps, err := GetAllMaps(driver)

		// assert
		assert.Nil(t, err)
		assert.Len(t, maps, 1)
	})

	mt.Run("failure-no-maps", func(mt *mtest.T) {
		// arrange for failure
		mt.AddMockResponses(
			mtest.CreateCursorResponse(
				0,
				mapSource,
				mtest.FirstBatch,
			),
		)
		// act
		driver := NewMockMongoDriver(mt.Client)
		maps, err := GetAllMaps(driver)

		// assert
		assert.Nil(t, maps)
		assert.Equal(t, err, errors.ErrMapNotFound)
	})
}

func TestGetPrimaryMapByUserID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mockMap := createMockMap([]byte("[]"))
	mt.Run("success", func(mt *mtest.T) {
		// arrange for success
		mt.AddMockResponses(
			mtest.CreateCursorResponse(
				1,
				mapSource,
				mtest.FirstBatch,
				createMapResponseData(mockMap),
			),
		)
		// act
		driver := NewMockMongoDriver(mt.Client)
		mapData, err := GetPrimaryMapByUserID(driver, mockMap.UserID)

		// assert
		assert.Nil(t, err)
		assert.NotNil(t, mapData)
	})

	mt.Run("failure-no-maps", func(mt *mtest.T) {
		// arrange for failure
		mt.AddMockResponses(
			mtest.CreateCursorResponse(
				0,
				mapSource,
				mtest.FirstBatch,
			),
		)
		// act
		driver := NewMockMongoDriver(mt.Client)
		_, err := GetPrimaryMapByUserID(driver, mockMap.UserID)

		// assert
		assert.NotNil(t, err)
		assert.Equal(t, err, errors.ErrMapNotFound)
	})
}

func TestGetMapByID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mockMap := createMockMap([]byte("[]"))
	mt.Run("success", func(mt *mtest.T) {
		// arrange for success
		mt.AddMockResponses(
			mtest.CreateCursorResponse(
				1,
				mapSource,
				mtest.FirstBatch,
				createMapResponseData(mockMap),
			),
		)
		// act
		driver := NewMockMongoDriver(mt.Client)
		mapData, err := GetMapByID(driver, mockMap.ID.Hex())

		// assert
		assert.Nil(t, err)
		assert.NotNil(t, mapData)
	})

	mt.Run("failure-no-maps", func(mt *mtest.T) {
		// arrange for failure
		mt.AddMockResponses(
			mtest.CreateCursorResponse(
				0,
				mapSource,
				mtest.FirstBatch,
			),
		)
		// act
		driver := NewMockMongoDriver(mt.Client)
		_, err := GetMapByID(driver, mockMap.ID.Hex())

		// assert
		assert.NotNil(t, err)
	})

	mt.Run("failure-invalid-id", func(mt *mtest.T) {
		// arrange for failure
		mt.AddMockResponses(
			mtest.CreateCursorResponse(
				0,
				mapSource,
				mtest.FirstBatch,
			),
		)
		// act
		driver := NewMockMongoDriver(mt.Client)
		_, err := GetMapByID(driver, "invalid-id")

		// assert
		assert.NotNil(t, err)
	})
}

func TestGetMapsByIDs(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mockMap := createMockMap([]byte("[]"))
	mt.Run("success", func(mt *mtest.T) {
		// arrange for success
		mt.AddMockResponses(
			mtest.CreateCursorResponse(
				1,
				mapSource,
				mtest.FirstBatch,
				createMapResponseData(mockMap),
			),
			CreateCursorEnd(mapSource),
		)
		// act
		driver := NewMockMongoDriver(mt.Client)
		mapData, err := GetMapsByIDs(driver, []string{mockMap.ID.Hex()})
		// assert
		assert.Nil(t, err)
		assert.NotNil(t, mapData)
	})

	mt.Run("failure-no-maps", func(mt *mtest.T) {
		// arrange for failure
		mt.AddMockResponses(
			mtest.CreateCursorResponse(
				0,
				mapSource,
				mtest.FirstBatch,
			),
			CreateCursorEnd(mapSource),
		)
		// act
		driver := NewMockMongoDriver(mt.Client)
		_, err := GetMapsByIDs(driver, []string{mockMap.ID.Hex()})

		// assert
		assert.Equal(t, err, errors.ErrMapNotFound)
	})
}

func TestGetMapByNameUserID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mockMap := createMockMap([]byte("[]"))
	mt.Run("success", func(mt *mtest.T) {
		// arrange for success
		mt.AddMockResponses(
			mtest.CreateCursorResponse(
				1,
				mapSource,
				mtest.FirstBatch,
				createMapResponseData(mockMap),
			),
		)
		// act
		driver := NewMockMongoDriver(mt.Client)
		mapData, err := GetMapByNameUserID(driver, mockMap.Name, mockMap.UserID)

		// assert
		assert.Nil(t, err)
		assert.NotNil(t, mapData)
		assert.Equal(t, mapData.Name, mockMap.Name)
		assert.Equal(t, mapData.UserID, mockMap.UserID)
	})

	mt.Run("failure-no-maps", func(mt *mtest.T) {
		// arrange for failure
		mt.AddMockResponses(
			mtest.CreateCursorResponse(
				0,
				mapSource,
				mtest.FirstBatch,
			),
		)
		// act
		driver := NewMockMongoDriver(mt.Client)
		_, err := GetMapByNameUserID(driver, mockMap.Name, mockMap.UserID)

		// assert
		assert.NotNil(t, err)
		assert.Equal(t, err, errors.ErrMapNotFound)
	})
}

func TestGetMapsByUserID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mockMap := createMockMap([]byte("[]"))
	mt.Run("success", func(mt *mtest.T) {
		// arrange for success
		mt.AddMockResponses(
			mtest.CreateCursorResponse(
				1,
				mapSource,
				mtest.FirstBatch,
				createMapResponseData(mockMap),
			),
			CreateCursorEnd(mapSource),
		)
		// act
		driver := NewMockMongoDriver(mt.Client)
		mapData, err := GetMapsByUserID(driver, mockMap.UserID)

		// assert
		assert.Nil(t, err)
		assert.NotNil(t, mapData)
		assert.Equal(t, mapData[0].UserID, mockMap.UserID)
	})

	mt.Run("failure-no-map", func(mt *mtest.T) {
		// arrange for failure
		mt.AddMockResponses(
			mtest.CreateCursorResponse(
				0,
				mapSource,
				mtest.FirstBatch,
			),
			CreateCursorEnd(mapSource),
		)
		// act
		driver := NewMockMongoDriver(mt.Client)
		_, err := GetMapsByUserID(driver, mockMap.UserID)

		// assert
		assert.NotNil(t, err)
		assert.Equal(t, err, errors.ErrMapNotFound)
	})
}

func TestUpdateMap(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mockMap := createMockMap("[]")
	mt.Run("success", func(mt *mtest.T) {
		// arrange for success
		mt.AddMockResponses(
			mtest.CreateCursorResponse(
				1,
				mapSource,
				mtest.FirstBatch,
				createMapResponseData(mockMap),
			),
			CreateCursorEnd(mapSource),
			SuccessResponse,
			SuccessResponse,
		)
		// act
		driver := NewMockMongoDriver(mt.Client)
		err := UpdateMap(driver, mockMap)

		// assert
		assert.Nil(t, err)
	})
}

func TestDeleteMap(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mockMap := createMockMap("[]")
	mt.Run("success", func(mt *mtest.T) {
		// arrange for success
		mt.AddMockResponses(
			SuccessResponse,
		)
		// act
		driver := NewMockMongoDriver(mt.Client)
		err := DeleteMap(driver, mockMap.ID.Hex())

		// assert
		assert.Nil(t, err)
	})

	mt.Run("failure-invalid-id", func(mt *mtest.T) {
		// arrange for failure
		mt.AddMockResponses(
			SuccessResponse,
		)
		// act
		driver := NewMockMongoDriver(mt.Client)
		err := DeleteMap(driver, "invalid-id")

		// assert
		assert.NotNil(t, err)
	})
}
