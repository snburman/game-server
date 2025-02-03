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
	Primary  bool               `json:"primary" bson:"primary"`
	Entrance struct {
		X int `json:"x" bson:"x"`
		Y int `json:"y" bson:"y"`
	}
	Portals []Portal `json:"portals" bson:"portals"`
	Data    T        `json:"data" bson:"data"`
}

// CreateMap creates a new map and reassigns primary map if necessary
func CreateMap(db DatabaseClient, m Map[string]) (primitive.ObjectID, error) {
	// check if map with the same name and userID exists
	_, err := db.GetOne(bson.M{"user_id": m.UserID, "name": m.Name}, mapsDBOptions)
	if err == nil {
		return primitive.NilObjectID, errors.ErrMapExists
	}

	// check if primary map already exists
	res, err := db.GetOne(bson.M{"user_id": m.UserID, "primary": true}, mapsDBOptions)
	if err == nil && m.Primary {
		var bm Map[[]byte]
		if err = utils.UnmarshalBSON(res, &bm); err != nil {
			return primitive.NilObjectID, errors.ErrCreatingMap
		}
		bm.Primary = false
		_, err = db.UpdateOne(bm.ID.Hex(), bm, mapsDBOptions)
		if err != nil {
			return primitive.NilObjectID, errors.ErrCreatingMap
		}
	} else if err != nil && !m.Primary {
		// if primary map does not exist, set primary to true
		m.Primary = true
	}

	// convert data to bytes
	byteMap := Map[[]byte]{
		UserID:   m.UserID,
		Name:     m.Name,
		Primary:  m.Primary,
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

// GetPrimaryMapByUserID retrieves the primary map by userID
func GetPrimaryMapByUserID(db DatabaseClient, userID string) (Map[[]PlayerAsset[PixelData]], error) {
	_map := *new(Map[[]PlayerAsset[PixelData]])
	res, err := db.GetOne(bson.M{"user_id": userID, "primary": true}, mapsDBOptions)
	if err != nil {
		return _map, errors.ErrMapNotFound
	}
	return unmarshalMap(res)
}

// GetMapByID retrieves a map by ID
func GetMapByID(db DatabaseClient, ID string) (Map[[]PlayerAsset[PixelData]], error) {
	empty := *new(Map[[]PlayerAsset[PixelData]])
	_id, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		return empty, err
	}

	res, err := db.GetOne(bson.M{"_id": _id}, mapsDBOptions)
	if err != nil {
		return empty, err
	}
	return unmarshalMap(res)
}

// GetMapByNameUserID retrieves a map by name and userID
func GetMapByNameUserID(db DatabaseClient, name, userID string) (Map[[]PlayerAsset[PixelData]], error) {
	_map := *new(Map[[]PlayerAsset[PixelData]])
	res, err := db.GetOne(bson.M{"user_id": userID, "name": name}, mapsDBOptions)
	if err != nil {
		return _map, errors.ErrMapNotFound
	}
	return unmarshalMap(res)
}

// GetMapsByUserID retrieves all maps by userID
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
		_map.Primary = bm.Primary
		_map.Entrance = bm.Entrance
		_map.Portals = bm.Portals
		maps = append(maps, *_map)
	}

	return maps, nil
}

// UpdateMap updates a map and reassigns primary map if necessary
func UpdateMap(db DatabaseClient, m Map[string]) error {
	if m.Primary {
		res, err := db.GetOne(bson.M{"user_id": m.UserID, "primary": true}, mapsDBOptions)
		if err == nil {
			// if primary map already exists, set it to false
			var bm Map[[]byte]
			if err = utils.UnmarshalBSON(res, &bm); err != nil {
				return errors.ErrUpdatingMap
			}
			bm.Primary = false
			_, err = db.UpdateOne(bm.ID.Hex(), bm, mapsDBOptions)
			if err != nil {
				return errors.ErrUpdatingMap
			}
		}
	}

	// convert data to bytes
	byteMap := Map[[]byte]{
		UserID:   m.UserID,
		Name:     m.Name,
		Primary:  m.Primary,
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

// unmarshalMap unmarshals map data from bytes to the correct type
func unmarshalMap(data any) (Map[[]PlayerAsset[PixelData]], error) {
	_map := *new(Map[[]PlayerAsset[PixelData]])
	var bm Map[[]byte]
	if err := utils.UnmarshalBSON(data, &bm); err != nil {
		return _map, errors.ErrMapWrongFormat
	}
	err := json.Unmarshal(bm.Data, &_map.Data)
	if err != nil {
		return _map, errors.ErrMapWrongFormat
	}
	_map.ID = bm.ID
	_map.UserID = bm.UserID
	_map.Name = bm.Name
	_map.Primary = bm.Primary
	_map.Entrance = bm.Entrance
	_map.Portals = bm.Portals

	return _map, nil
}
