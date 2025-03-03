package db

import (
	"testing"

	"github.com/snburman/game-server/errors"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

var userSource = "game.users"

func createMockUser() User {
	return User{
		UserName: "username",
		Password: "passwordABC123",
	}
}

func createUserResponseData(u User) bson.D {
	return bson.D{
		{Key: "_id", Value: u.ID},
		{Key: "username", Value: u.UserName},
		{Key: "password", Value: u.Password},
		{Key: "role", Value: u.Role},
		{Key: "banned", Value: u.Banned},
	}
}

func TestCreateUser(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mockUser := createMockUser()

	mt.Run("success", func(mt *mtest.T) {
		// arrange for success
		mt.AddMockResponses(
			SuccessResponse,
		)

		// act
		driver := NewMockMongoDriver(mt.Client)
		_, err := CreateUser(driver, mockUser)

		// assert
		assert.Nil(t, err)
	})

	mt.Run("failure-weak-password", func(mt *mtest.T) {
		_mockUser := mockUser
		_mockUser.Password = "password"
		// arrange for failure
		mt.AddMockResponses(
			mtest.CreateCursorResponse(
				0,
				userSource,
				mtest.FirstBatch,
			),
		)

		// act
		driver := NewMockMongoDriver(mt.Client)
		_, err := CreateUser(driver, _mockUser)

		// assert
		assert.NotNil(t, err)
		assert.Equal(t, err, errors.ErrWeakPassword)
	})
}

func TestGetUserByID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mockUser := createMockUser()

	mt.Run("success", func(mt *mtest.T) {
		// arrange for success
		mt.AddMockResponses(
			mtest.CreateCursorResponse(
				1,
				userSource,
				mtest.FirstBatch,
				createUserResponseData(mockUser),
			),
			CreateCursorEnd(userSource),
		)

		// act
		driver := NewMockMongoDriver(mt.Client)
		res, err := GetUserByID(driver, mockUser.ID.Hex())

		// assert
		assert.Nil(t, err)
		assert.Equal(t, res, mockUser)
	})

	mt.Run("failure-user-not-found", func(mt *mtest.T) {
		// arrange for failure
		mt.AddMockResponses(
			mtest.CreateCursorResponse(
				0,
				userSource,
				mtest.FirstBatch,
			),
		)

		// act
		driver := NewMockMongoDriver(mt.Client)
		res, err := GetUserByID(driver, mockUser.ID.Hex())

		// assert
		assert.NotNil(t, err)
		assert.Equal(t, res, User{})
	})
}

func TestGetUserByUserName(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mockUser := createMockUser()

	mt.Run("success", func(mt *mtest.T) {
		// arrange for success
		mt.AddMockResponses(
			mtest.CreateCursorResponse(
				1,
				userSource,
				mtest.FirstBatch,
				createUserResponseData(mockUser),
			),
			CreateCursorEnd(userSource),
		)

		// act
		driver := NewMockMongoDriver(mt.Client)
		res, err := GetUserByUserName(driver, mockUser.UserName)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, res, mockUser)
	})

	mt.Run("failure-user-not-found", func(mt *mtest.T) {
		// arrange for failure
		mt.AddMockResponses(
			mtest.CreateCursorResponse(
				0,
				userSource,
				mtest.FirstBatch,
			),
		)

		// act
		driver := NewMockMongoDriver(mt.Client)
		res, err := GetUserByUserName(driver, mockUser.UserName)

		// assert
		assert.NotNil(t, err)
		assert.Equal(t, res, User{})
	})
}

func TestDeleteUser(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mockUser := createMockUser()

	mt.Run("success", func(mt *mtest.T) {
		// arrange for success
		mt.AddMockResponses(
			SuccessResponse,
		)

		// act
		driver := NewMockMongoDriver(mt.Client)
		_, err := DeleteUser(driver, mockUser.ID.Hex())

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
		_, err := DeleteUser(driver, "invalid-id")

		// assert
		assert.NotNil(t, err)
	})
}

func TestUpdateUser(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mockUser := createMockUser()

	mt.Run("success", func(mt *mtest.T) {
		// arrange for success
		mt.AddMockResponses(
			SuccessResponse,
		)

		// act
		driver := NewMockMongoDriver(mt.Client)
		err := UpdateUser(driver, mockUser)

		// assert
		assert.Nil(t, err)
	})

	mt.Run("failure-weak-password", func(mt *mtest.T) {
		_mockUser := mockUser
		_mockUser.Password = "password"
		// arrange for failure
		mt.AddMockResponses(
			SuccessResponse,
		)

		// act
		driver := NewMockMongoDriver(mt.Client)
		err := UpdateUser(driver, _mockUser)

		// assert
		assert.NotNil(t, err)
		assert.Equal(t, err, errors.ErrWeakPassword)
	})
}
