package db

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type User struct {
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
}

func GetUser(db *mongo.Client, username string) (User, error) {
	mdb := db.Database("magic_game")
	res := mdb.Collection("users").FindOne(context.Background(), bson.M{
		"username": username,
	})

	user := User{}
	err := res.Decode(&user)
	return user, err
}

func CreateUser(db *mongo.Client, user User) (*mongo.InsertOneResult, error) {
	mdb := db.Database("magic_game")
	return mdb.Collection("users").InsertOne(context.Background(), user)
}

func DeleteUser(db *mongo.Client, username string) (*mongo.DeleteResult, error) {
	mdb := db.Database("magic_game")
	return mdb.Collection("users").DeleteOne(context.Background(), bson.M{
		"username": username,
	})
}
