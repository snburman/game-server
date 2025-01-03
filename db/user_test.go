package db

import (
	"strconv"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func newUserFixtures() map[string]any {
	fixtures := make(map[string]any)
	for x := 123; x < 200; x++ {
		data := bson.M{"user_id": strconv.Itoa(x)}
		b, err := bson.Marshal(data)
		if err != nil {
			panic("failed to init fixtures")
		}
		fixtures[string(b)] = data
	}
	return fixtures
}

func TestNewUser(t *testing.T) {
	// id := "999"
	// username := "CAPITALIZED"
	// user := User{
	// 	ID:       id,
	// 	UserName: username,
	// }
	// profile := NewUser(user)
	// if profile.UserName != strings.ToLower(username) {
	// 	t.Errorf("expected %v, got %v", username, profile.UserName)
	// }
}

func TestCreateUserProfile(t *testing.T) {
	// client := NewMockMongoClient(newUserFixtures())
	// id := "0"
	// up := UserProfile{UserID: id}
	// _, err := CreateUserProfile(client, up)
	// if err != nil {
	// 	t.Error(err)
	// }
	// b, err := bson.Marshal(up)
	// doc, ok := client.Fixtures[string(b)]
	// if !ok {
	// 	t.Errorf("expected fixture, got %v", doc)
	// }
	// profile, ok := doc.(UserProfile)
	// if !ok {
	// 	t.Errorf("expected UserProfile type")
	// }
	// if profile.UserID != up.UserID {
	// 	t.Errorf("expected %v, got %v", profile.UserID, up.UserID)
	// }
}

func TestGetUserProfileByID(t *testing.T) {
	// client := NewMockMongoClient(newUserFixtures())

}

func TestDeleteUserProfle(t *testing.T) {

}
