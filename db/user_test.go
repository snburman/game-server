package db

import (
	"strconv"
	"strings"
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

func TestNewUserProfile(t *testing.T) {
	id := "999"
	username := "CAPITALIZED"
	user := User{
		ID:       id,
		UserName: username,
	}
	profile := NewUserProfile(user)
	if profile.UserID != id {
		t.Errorf("expected %v, got %v", id, profile.UserID)
	}
	if profile.UserName != strings.ToLower(username) {
		t.Errorf("expected %v, got %v", username, profile.UserName)
	}
}

func TestCreateUserProfile(t *testing.T) {
	client := NewMockMongoClient(newUserFixtures())
	id := "0"
	up := UserProfile{UserID: id}
	err := CreateUserProfile(client, up)
	if err != nil {
		t.Error(err)
	}
	b, err := bson.Marshal(up)
	doc, ok := client.Fixtures[string(b)]
	if !ok {
		t.Errorf("expected fixture, got %v", doc)
	}
	profile, ok := doc.(UserProfile)
	if !ok {
		t.Errorf("expected UserProfile type")
	}
	if profile.UserID != up.UserID {
		t.Errorf("expected %v, got %v", profile.UserID, up.UserID)
	}
}

func TestGetUserProfileByID(t *testing.T) {
	client := NewMockMongoClient(newUserFixtures())
	t.Run("get profile correct id", func(t *testing.T) {
		id := "123"
		profile, err := GetUserProfileByID(client, id)
		if err != nil {
			t.Error(err)
		}
		if profile.UserID != id {
			t.Error("incorrect id")
		}
	})

	t.Run("get profile incorrect id", func(t *testing.T) {
		id := "1"
		_, err := GetUserProfileByID(client, id)
		if err == nil {
			t.Error("expected error")
		}
	})
}

func TestDeleteUserProfle(t *testing.T) {
	client := NewMockMongoClient(newUserFixtures())
	initial := len(client.Fixtures)
	count, err := DeleteUserProfile(client, "123")
	if err != nil {
		t.Error(err)
	}
	if count != 1 {
		t.Errorf("expected %v, got %v", 1, count)
	}
	after := len(client.Fixtures)
	if initial-count != after {
		t.Errorf("expected %v, got %v", initial-count, after)
	}
}
