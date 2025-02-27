package db

import (
	"context"
	"encoding/json"
	"log"

	"github.com/snburman/game-server/config"
	"github.com/snburman/game-server/errors"
	"github.com/snburman/game-server/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var assetDBOptions DatabaseClientOptions = DatabaseClientOptions{
	Database: GameDatabase,
	Table:    PlayerImagesCollection,
}

type Pixel struct {
	X     int    `json:"x"`
	Y     int    `json:"y"`
	R     int    `json:"r"`
	G     int    `json:"g"`
	B     int    `json:"b"`
	A     int    `json:"a"`
	Color string `json:"color"`
}

type PixelData = [][]Pixel

type AssetType string

const (
	ASSET_TILE         AssetType = "tile"
	ASSET_OBJECT       AssetType = "object"
	ASSET_PORTAL       AssetType = "portal"
	ASSET_PLAYER_UP    AssetType = "player_up"
	ASSET_PLAYER_DOWN  AssetType = "player_down"
	ASSET_PLAYER_LEFT  AssetType = "player_left"
	ASSET_PLAYER_RIGHT AssetType = "player_right"
)

type PlayerAsset[T any] struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	UserID    string             `json:"user_id" bson:"user_id"`
	Name      string             `json:"name" bson:"name"`
	AssetType AssetType          `json:"asset_type" bson:"asset_type"`
	X         int                `json:"x" bson:"x"`
	Y         int                `json:"y" bson:"y"`
	Width     int                `json:"width" bson:"width"`
	Height    int                `json:"height" bson:"height"`
	Data      T                  `json:"data" bson:"data"`
}

// CreatePlayerAsset will return an error if 'p' already exists.
// Stores PlayerAsset[[]byte] in db
func CreatePlayerAsset(db *MongoDriver, p PlayerAsset[string]) (primitive.ObjectID, error) {
	// check if asset with same name and userID exists
	_, err := db.GetOne(bson.M{"user_id": p.UserID, "name": p.Name}, assetDBOptions)
	if err == nil {
		return primitive.NilObjectID, errors.ErrImageExists
	}

	// convert to byte asset
	byteAsset := PlayerAsset[[]byte]{
		UserID:    p.UserID,
		Name:      p.Name,
		AssetType: p.AssetType,
		X:         p.X,
		Y:         p.Y,
		Width:     p.Width,
		Height:    p.Height,
		Data:      []byte(p.Data),
	}

	id, err := db.CreateOne(byteAsset, assetDBOptions)
	if err != nil {
		return primitive.NilObjectID, err
	}

	insertedID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return insertedID, nil
}

func GetPlayerAssetsByUserID(db *MongoDriver, userID string) ([]PlayerAsset[PixelData], error) {
	assets := []PlayerAsset[PixelData]{}

	// get assets with byte data
	byteAssets := []PlayerAsset[[]byte]{}
	err := db.Get(bson.M{"user_id": userID}, assetDBOptions, &byteAssets)
	if err != nil {
		return assets, err
	}

	for _, img := range byteAssets {
		_img := new(PlayerAsset[PixelData])
		// decode the json string
		err := json.Unmarshal(img.Data, &_img.Data)
		if err != nil {
			log.Println("error decoding image: ", err)
			return assets, err
		}
		_img.ID = img.ID
		_img.UserID = img.UserID
		_img.Name = img.Name
		_img.AssetType = img.AssetType
		_img.Width = img.Width
		_img.Height = img.Height
		assets = append(assets, *_img)
	}

	return assets, nil
}

func GetPlayerCharactersByUserIDs(db *MongoDriver, userIDs []string) ([]PlayerAsset[PixelData], error) {
	filter := bson.A{
		bson.D{{
			Key: "$match", Value: bson.D{
				{Key: "user_id", Value: bson.D{{Key: "$in", Value: userIDs}}},
				{Key: "$or", Value: bson.A{
					bson.D{{Key: "asset_type", Value: ASSET_PLAYER_UP}},
					bson.D{{Key: "asset_type", Value: ASSET_PLAYER_DOWN}},
					bson.D{{Key: "asset_type", Value: ASSET_PLAYER_LEFT}},
					bson.D{{Key: "asset_type", Value: ASSET_PLAYER_RIGHT}},
				}},
			},
		}},
	}
	assets := []PlayerAsset[PixelData]{}
	// get assets with byte data
	byteAssets := []PlayerAsset[[]byte]{}
	cursor, err := db.Client.
		Database(assetDBOptions.Database).
		Collection(assetDBOptions.Table).
		Aggregate(context.Background(), filter)
	if err != nil {
		return assets, err
	}
	defer cursor.Close(context.Background())
	err = cursor.All(context.Background(), &byteAssets)
	if err != nil {
		return assets, err
	}

	// unmarshal data from []byte to PixelData
	for _, img := range byteAssets {
		_img := new(PlayerAsset[PixelData])
		// decode the json string
		err := json.Unmarshal(img.Data, &_img.Data)
		if err != nil {
			log.Println("error decoding image: ", err)
			return assets, err
		}
		_img.ID = img.ID
		_img.UserID = img.UserID
		_img.Name = img.Name
		_img.AssetType = img.AssetType
		_img.Width = img.Width
		_img.Height = img.Height
		assets = append(assets, *_img)
	}

	return assets, nil
}

// AppendMapPlayerCharacter gets all character assets for a user and appends them to the map
func AppendMapPlayerCharacter(db *MongoDriver, userID string, _map Map[[]PlayerAsset[PixelData]]) (Map[[]PlayerAsset[PixelData]], error) {
	// add character assets
	charAssets, err := GetPlayerCharactersByUserIDs(db, []string{userID})
	if err != nil {
		return _map, errors.ServerError(err.Error())
	}
	// set character assets x,y to entrance
	for i := range charAssets {
		charAssets[i].X = _map.Entrance.X
		charAssets[i].Y = _map.Entrance.Y
	}
	_map.Data = append(_map.Data, charAssets...)
	return _map, nil
}

func GetPlayerAssetByNameUserID(db *MongoDriver, name string, userID string) (PlayerAsset[PixelData], error) {
	asset := PlayerAsset[PixelData]{}
	res, err := db.GetOne(bson.M{"name": name, "user_id": userID}, assetDBOptions)
	if err != nil {
		return asset, errors.ErrImageNotFound
	}

	var byteAsset PlayerAsset[[]byte]
	if err = utils.UnmarshalBSON(res, &byteAsset); err != nil {
		return asset, err
	}

	// unmarshal data field from []byte to PixelData
	if err = json.Unmarshal(byteAsset.Data, &asset.Data); err != nil {
		return asset, errors.ErrImageWrongFormat
	}
	// copy properties
	asset.ID = byteAsset.ID
	asset.UserID = byteAsset.UserID
	asset.Name = byteAsset.Name
	asset.AssetType = byteAsset.AssetType
	asset.X = byteAsset.X
	asset.Y = byteAsset.Y
	asset.Width = byteAsset.Width
	asset.Height = byteAsset.Height

	return asset, nil
}

func GetDefaultPlayerCharacter(db *MongoDriver) (PlayerAsset[PixelData], error) {
	return GetPlayerAssetByNameUserID(
		db, "default_character", config.Env().ADMIN_ID,
	)
}

func UpdatePlayerAsset(db *MongoDriver, p PlayerAsset[string]) error {
	// convert to byte asset
	byteAsset := PlayerAsset[[]byte]{
		UserID:    p.UserID,
		Name:      p.Name,
		AssetType: p.AssetType,
		X:         p.X,
		Y:         p.Y,
		Width:     p.Width,
		Height:    p.Height,
		Data:      []byte(p.Data),
	}
	_, err := db.UpdateOne(p.ID.Hex(), byteAsset, assetDBOptions)
	return err
}

func DeletePlayerAsset(db *MongoDriver, id string) (count int, err error) {
	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return 0, err
	}
	return db.Delete(bson.M{"_id": _id}, assetDBOptions)
}
