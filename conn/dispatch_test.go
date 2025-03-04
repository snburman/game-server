package conn

import (
	"encoding/json"
	"testing"

	"github.com/snburman/game-server/db"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestNewDispatch(t *testing.T) {
	conn := NewMockConn()
	dispatch := NewDispatch("123", conn, UpdatePlayer, "test")
	assert.Equal(t, UpdatePlayer, dispatch.Function)
	assert.Equal(t, "test", dispatch.Data)
	assert.Equal(t, conn, dispatch.conn)
	assert.Equal(t, "123", dispatch.ID)
}

func TestMarshalDispatch(t *testing.T) {
	conn := NewMockConn()
	dispatch := NewDispatch("123", conn, UpdatePlayer, "test")
	marshalled := dispatch.Marshal()
	assert.Equal(t, dispatch.ID, marshalled.ID)
	assert.Equal(t, dispatch.Function, marshalled.Function)
	assert.Equal(t, dispatch.conn, marshalled.conn)
	assert.Equal(t, []byte(`"test"`), marshalled.Data)
}

func TestParseDispatch(t *testing.T) {
	conn := NewMockConn()
	playerUpdate := PlayerUpdate{
		UserID: "123",
		MapID:  "456",
	}

	dispatch := NewDispatch("123", conn, UpdatePlayer, playerUpdate)
	marshalled := dispatch.Marshal()
	parsed := ParseDispatch[PlayerUpdate](marshalled)
	assert.Equal(t, dispatch.ID, parsed.ID)
	assert.Equal(t, dispatch.Function, parsed.Function)
	assert.Equal(t, dispatch.conn, parsed.conn)
	assert.Equal(t, dispatch.Data, parsed.Data)
	assert.Equal(t, playerUpdate, parsed.Data)
}

func TestPublishDispatch(t *testing.T) {
	conn := NewMockConn()
	dispatch := NewDispatch("123", conn, UpdatePlayer, "test")
	dispatch.Publish()
	msg := <-conn.Messages
	assert.Equal(
		t,
		`{"id":"123","function":"update_player","data":"test"}`,
		string(msg),
	)
}

func TestRouteDispatch(t *testing.T) {
	conn := NewMockConn()
	playerUpdate := PlayerUpdate{
		UserID: "123",
		MapID:  "456",
	}
	dispatch := NewDispatch("123", conn, UpdatePlayer, playerUpdate)
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("new-player", func(mt *mtest.T) {
		db.MongoDB = db.NewMockMongoDriver(mt.Client)
		// arrange
		mt.AddMockResponses(
			mtest.CreateCursorResponse(
				1,
				"game.player_images",
				mtest.FirstBatch,
				db.CreatePlayerAssetResponseData(db.CreateMockPlayerAsset("[]")),
			),
			db.CreateCursorEnd("game.player_images"),
		)

		// act
		RouteDispatch(dispatch.Marshal())
		msg := <-conn.Messages
		var d Dispatch[[]byte]
		err := json.Unmarshal(msg, &d)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, UpdatePlayer, dispatch.Function)
	})

	mt.Run("new-map-id", func(mt *mtest.T) {
		db.MongoDB = db.NewMockMongoDriver(mt.Client)
		mt.AddMockResponses(
			mtest.CreateCursorResponse(
				1,
				"game.player_images",
				mtest.FirstBatch,
				db.CreatePlayerAssetResponseData(db.CreateMockPlayerAsset("[]")),
			),
			db.CreateCursorEnd("game.player_images"),
		)

		// act
		conn.MapID = "789"
		RouteDispatch(dispatch.Marshal())
		msg := <-conn.Messages
		var d Dispatch[[]byte]
		err := json.Unmarshal(msg, &d)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, UpdatePlayer, dispatch.Function)
		assert.Equal(t, "456", conn.MapID)
	})
}
