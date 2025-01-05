package db

import (
	"errors"
	"strings"

	"github.com/snburman/game_server/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var userDBOptions DatabaseClientOptions = DatabaseClientOptions{
	Database: GameDatabase,
	Table:    UserProfilesCollection,
}

type Role string

const CreatorRole Role = "creator"
const PlayerRole Role = "player"

type User struct {
	ID       primitive.ObjectID     `json:"_id,omitempty" bson:"_id,omitempty"`
	UserName string                 `json:"username,omitempty" bson:"username"`
	Password string                 `json:"password,omitempty" bson:"password"`
	Role     Role                   `json:"role" bson:"role"`
	Worlds   map[string]interface{} `json:"worlds" bson:"worlds"`
	Banned   bool                   `json:"banned" bson:"banned"`
}

func CreateUser(db DatabaseClient, u User) (instertedID primitive.ObjectID, err error) {
	user := User{
		UserName: strings.ToLower(u.UserName),
		Worlds:   make(map[string]interface{}),
	}
	password, err := utils.HashPassword(u.Password)
	if err != nil {
		return primitive.NilObjectID, err
	}
	user.Password = password

	id, err := db.CreateOne(user, userDBOptions)
	if err != nil {
		return primitive.NilObjectID, err
	}
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID, errors.New("error generating ObjectID")
	}
	return objectID, err
}

func GetUserByID(db DatabaseClient, userID string) (User, error) {
	_id, err := primitive.ObjectIDFromHex(userID)
	res, err := db.GetOne(bson.M{"_id": _id}, userDBOptions)
	if err != nil {
		return User{}, err
	}
	b, err := bson.Marshal(res)
	var user User
	err = bson.Unmarshal(b, &user)
	if err != nil {
		return User{}, errors.New("error umarshalling user")
	}
	return user, nil
}

func GetUserByUserName(db DatabaseClient, userName string) (User, error) {
	res, err := db.GetOne(bson.M{"username": userName}, userDBOptions)
	if err != nil {
		return User{}, err
	}
	b, err := bson.Marshal(res)
	var user User
	err = bson.Unmarshal(b, &user)
	if err != nil {
		return User{}, errors.New("error umarshalling user")
	}
	return user, nil
}

func DeleteUser(db DatabaseClient, userID string) (count int, err error) {
	_id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return 0, err
	}
	return db.Delete(bson.M{"_id": _id}, userDBOptions)
}

func UpdateUser(db DatabaseClient, u User) error {
	if u.Password != "" {
		password, err := utils.HashPassword(u.Password)
		if err != nil {
			return err
		}
		u.Password = password
	}
	_, err := db.UpdateOne(u.ID.Hex(), u, userDBOptions)
	return err
}
