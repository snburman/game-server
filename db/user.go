package db

import (
	"errors"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var userDBOptions DatabaseClientOptions = DatabaseClientOptions{
	Database: GameDatabase,
	Table:    UserProfilesCollection,
}

type Role string

const CreatorRole Role = "creator"

type User struct {
	ID       string `json:"user_id"`
	Email    string `json:"email"`
	UserName string `json:"username"`
	Password string `json:"password,omitempty"`
}

type UserMetaData map[string]struct {
	UserName string `json:"username"`
}

type UserProfile struct {
	ID       primitive.ObjectID     `json:"-" bson:"_id"`
	UserID   string                 `json:"user_id" bson:"user_id"`
	UserName string                 `json:"username,omitempty" bson:"username"`
	Role     Role                   `json:"role" bson:"role"`
	Worlds   map[string]interface{} `json:"worlds" bson:"worlds"`
}

func NewUserProfile(u User) UserProfile {
	userName := strings.ToLower(u.UserName)
	return UserProfile{
		UserID:   u.ID,
		UserName: userName,
		Worlds:   make(map[string]interface{}),
	}
}

func CreateUserProfile(db DatabaseClient, up UserProfile) error {
	return db.CreateOne(up, userDBOptions)
}

func GetUserProfileByID(db DatabaseClient, userID string) (UserProfile, error) {
	res, err := db.GetOne(bson.M{"user_id": userID}, userDBOptions)
	if err != nil {
		return UserProfile{}, err
	}
	b, err := bson.Marshal(res)
	var profile UserProfile
	err = bson.Unmarshal(b, &profile)
	if err != nil {
		return UserProfile{}, errors.New("error_marshalling_profile")
	}
	return profile, nil
}

func DeleteUserProfile(db DatabaseClient, userID string) (int, error) {
	return db.Delete(bson.M{"user_id": userID}, userDBOptions)
}
