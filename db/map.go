package db

import (
	"encoding/json"
	"log"

	"github.com/snburman/game_server/errors"
	"github.com/snburman/game_server/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var mapsDBOptions = DatabaseClientOptions{
	Database: GameDatabase,
	Table:    PlayerMapsCollection,
}

type Portal struct {
	MapID string `json:"map_id" bson:"map_id"`
	X     int    `json:"x" bson:"x"`
	Y     int    `json:"y" bson:"y"`
}

type Map[T any] struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	UserID   string             `json:"user_id" bson:"user_id"`
	Name     string             `json:"name" bson:"name"`
	Entrance struct {
		X int `json:"x" bson:"x"`
		Y int `json:"y" bson:"y"`
	}
	Portals []Portal `json:"portals" bson:"portals"`
	Data    T        `json:"data" bson:"data"`
}

func CreateMap(db DatabaseClient, m Map[string]) (primitive.ObjectID, error) {
	// check if map with the same name and userID exists
	_, err := db.GetOne(bson.M{"user_id": m.UserID, "name": m.Name}, mapsDBOptions)
	if err == nil {
		return primitive.NilObjectID, errors.ErrMapExists
	}

	// convert data to bytes
	byteMap := Map[[]byte]{
		UserID:   m.UserID,
		Name:     m.Name,
		Entrance: m.Entrance,
		Portals:  m.Portals,
		Data:     []byte(m.Data),
	}

	id, err := db.CreateOne(byteMap, mapsDBOptions)
	if err != nil {
		return primitive.NilObjectID, err
	}

	insertedID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return insertedID, nil
}

func GetMapByID(db DatabaseClient, ID string) (Map[[]PlayerAsset[PixelData]], error) {
	_map := *new(Map[[]PlayerAsset[PixelData]])
	_id, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		return _map, err
	}

	res, err := db.GetOne(bson.M{"_id": _id}, mapsDBOptions)
	if err != nil {
		return _map, err
	}
	var bm Map[[]byte]
	if err = utils.UnmarshalBSON(res, &bm); err != nil {
		return _map, err
	}
	err = json.Unmarshal(bm.Data, &_map.Data)
	if err != nil {
		return _map, err
	}
	_map.ID = bm.ID
	_map.UserID = bm.UserID
	_map.Name = bm.Name
	_map.Entrance = bm.Entrance
	_map.Portals = bm.Portals

	return _map, nil
}

func GetMapByNameUserID(db DatabaseClient, name, userID string) (Map[[]PlayerAsset[PixelData]], error) {
	_map := *new(Map[[]PlayerAsset[PixelData]])
	res, err := db.GetOne(bson.M{"user_id": userID, "name": name}, mapsDBOptions)
	if err != nil {
		return _map, errors.ErrMapNotFound
	}

	var bm Map[[]byte]
	if err = utils.UnmarshalBSON(res, &bm); err != nil {
		return _map, err
	}

	// unmarshal data field from []byte to PixelData
	if err = json.Unmarshal(bm.Data, &_map.Data); err != nil {
		return _map, errors.ErrMapWrongFormat
	}
	// copy properties
	_map.ID = bm.ID
	_map.UserID = bm.UserID
	_map.Name = bm.Name
	_map.Entrance = bm.Entrance
	_map.Portals = bm.Portals

	return _map, nil
}

func GetMapsByUserID(db DatabaseClient, userID string) ([]Map[[]PlayerAsset[PixelData]], error) {
	var byteMaps []Map[[]byte]
	err := db.Get(bson.M{"user_id": userID}, mapsDBOptions, &byteMaps)
	if err != nil {
		return nil, errors.ErrMapNotFound
	}

	var maps []Map[[]PlayerAsset[PixelData]]
	for _, bm := range byteMaps {
		_map := new(Map[[]PlayerAsset[PixelData]])
		err := json.Unmarshal(bm.Data, &_map.Data)
		if err != nil {
			log.Println("error decoding map images: ", err)
			return nil, err
		}
		_map.ID = bm.ID
		_map.UserID = bm.UserID
		_map.Name = bm.Name
		_map.Entrance = bm.Entrance
		_map.Portals = bm.Portals
		maps = append(maps, *_map)
	}

	return maps, nil
}

func UpdateMap(db DatabaseClient, m Map[string]) error {
	// convert data to bytes
	byteMap := Map[[]byte]{
		UserID:   m.UserID,
		Name:     m.Name,
		Entrance: m.Entrance,
		Portals:  m.Portals,
		Data:     []byte(m.Data),
	}

	_, err := db.UpdateOne(m.ID.Hex(), byteMap, mapsDBOptions)
	if err != nil {
		return err
	}
	return nil
}

func DeleteMap(db DatabaseClient, ID string) error {
	_id, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		return err
	}

	_, err = db.Delete(bson.M{"_id": _id}, mapsDBOptions)
	if err != nil {
		return err
	}
	return nil
}
