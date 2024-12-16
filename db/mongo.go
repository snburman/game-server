package db

import (
	"context"
	"log"
	"reflect"
	"time"

	"github.com/snburman/game_server/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Databases
const GameDatabase = "game"

// Collections
const UserProfilesCollection = "user_profiles"
const ImagesCollection = "images"
const PlayerImagesCollection = "player_images"

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

func (m *MongoDriver) Get(params any, opts DatabaseClientOptions, dest *[]any) error {
	mdb := MongoDB.Client.Database(opts.Database)
	ctx := context.Background()
	res, err := mdb.Collection(opts.Table).Find(ctx, params)
	if err == nil {
		res.All(ctx, dest)
	}
	return err
}

func (m *MongoDriver) GetOne(params any, opts DatabaseClientOptions) (any, error) {
	mdb := MongoDB.Client.Database(opts.Database)
	res := mdb.Collection(opts.Table).FindOne(context.Background(), params)
	var dest any
	err := res.Decode(&dest)
	return dest, err
}

func (m *MongoDriver) CreateOne(document any, opts DatabaseClientOptions) error {
	mdb := MongoDB.Client.Database(opts.Database)
	_, err := mdb.Collection(opts.Table).InsertOne(context.Background(), document)
	return err
}

func (m *MongoDriver) UpdateOne(document any, opts DatabaseClientOptions) (any, error) {
	return nil, nil
}

func (m *MongoDriver) Delete(params any, opts DatabaseClientOptions) (count int, err error) {
	mdb := MongoDB.Client.Database(opts.Database)
	res, err := mdb.Collection(opts.Table).DeleteOne(context.Background(), params)
	return int(res.DeletedCount), err
}
