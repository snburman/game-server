package db

import "testing"

var mockPlayerAsset = PlayerAsset[string]{
	UserID:    "67bfa82f165e6e4169699147",
	Name:      "test_image",
	AssetType: ASSET_TILE,
	X:         0,
	Y:         0,
	Width:     16,
	Height:    16,
}

func TestCreatePlayerAsset(t *testing.T) {
	name := "CreatePlayerAsset-"
	// create a new player asset
	t.Run(name+"-success", func(t *testing.T) {
		client := NewMockDatabaseClient[PlayerAsset[string]]()
		_, err := CreatePlayerAsset(client, mockPlayerAsset)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})
	// client.AddData(assetDBOptions.Table, mockPlayerAsset)
}
