package db

import (
	"context"
	"log"
	"reflect"
	"time"

	"github.com/snburman/magic_game_server/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoDB *MongoDriver

type MongoDriver struct {
	Client *mongo.Client
}

func NewMongoDriver() {
	if MongoDB != nil {
		return
	}
	md := &MongoDriver{}
	if err := md.Connect(); err != nil {
		log.Panicln(err.Error())
	}
	MongoDB = md
}

func (m *MongoDriver) Connect() error {
	uri := config.Env().MONGO_URI
	t := reflect.TypeOf(bson.M{})
	reg := bson.NewRegistry()
	reg.RegisterTypeMapEntry(bson.TypeEmbeddedDocument, t)
	client, err := mongo.Connect(
		context.TODO(),
		options.Client().ApplyURI(uri).SetTimeout(time.Second*30).SetRegistry(reg),
	)
	if err != nil {
		return err
	}
	m.Client = client
	log.Println("Connected to MongoDB...")
	return nil
}

// Disconnect() should be defered after calling Connect()
func (m *MongoDriver) Disconnect() error {
	if err := m.Client.Disconnect(context.TODO()); err != nil {
		return err
	}
	log.Println("Disconnected from MongoDB...")
	return nil
}
