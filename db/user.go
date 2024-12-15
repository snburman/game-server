package db

import (
	"context"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

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
	UserID   string                 `json:"user_id" bson:"user_id"`
	UserName string                 `json:"username,omitempty"`
	Role     Role                   `json:"role"`
	Worlds   map[string]interface{} `json:"worlds"`
}

func NewUserProfile(u User) UserProfile {
	userName := strings.ToLower(u.UserName)
	return UserProfile{
		UserID:   u.ID,
		UserName: userName,
		Worlds:   make(map[string]interface{}),
	}
}

func CreateUserProfile(db *mongo.Client, up UserProfile) (*mongo.InsertOneResult, error) {
	mdb := db.Database(GameDatabase)
	return mdb.Collection(UserProfilesCollection).InsertOne(context.Background(), up)
}

func GetUserProfileByID(db *mongo.Client, ID string) (UserProfile, error) {
	mdb := db.Database(GameDatabase)
	res := mdb.Collection(UserProfilesCollection).FindOne(context.Background(), bson.M{
		"user_id": ID,
	})

	up := UserProfile{}
	err := res.Decode(&up)
	return up, err
}

func DeleteUserProfile(db *mongo.Client, userID string) (*mongo.DeleteResult, error) {
	mdb := db.Database(GameDatabase)
	return mdb.Collection(UserProfilesCollection).DeleteOne(context.Background(), bson.M{
		"user_id": userID,
	})
}
