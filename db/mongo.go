package db

import (
	"context"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/snburman/game-server/config"
	"github.com/snburman/game-server/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Databases
const GameDatabase = "game"

// Collections
const UserProfilesCollection = "user_profiles"
const ImagesCollection = "images"
const PlayerImagesCollection = "player_images"
const PlayerMapsCollection = "player_maps"

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

func (m *MongoDriver) Get(params any, opts DatabaseClientOptions, dest any) error {
	mdb := m.Client.Database(opts.Database)
	ctx := context.Background()
	res, err := mdb.Collection(opts.Table).Find(ctx, params)
	if err == nil {
		res.All(ctx, dest)
	}
	return err
}

func (m *MongoDriver) GetOne(params any, opts DatabaseClientOptions) (any, error) {
	mdb := m.Client.Database(opts.Database)
	res := mdb.Collection(opts.Table).FindOne(context.Background(), params)
	var dest any
	err := res.Decode(&dest)
	return dest, err
}

func (m *MongoDriver) CreateOne(document any, opts DatabaseClientOptions) (insertedID string, err error) {
	mdb := m.Client.Database(opts.Database)
	res, err := mdb.Collection(opts.Table).InsertOne(context.Background(), document)
	id := res.InsertedID.(primitive.ObjectID)
	return id.Hex(), err
}

func (m *MongoDriver) UpdateOne(id string, document any, opts DatabaseClientOptions) (any, error) {
	mdb := m.Client.Database(opts.Database)
	var updates bson.D

	typeData := reflect.TypeOf(document)
	values := reflect.ValueOf(document)

	// https://joshua-etim.medium.com/how-i-update-documents-in-mongodb-with-golang-94485dbe54f7
	for i := 1; i < typeData.NumField(); i++ {
		field := typeData.Field(i)
		val := values.Field(i)
		tag := field.Tag.Get("bson")
		// disregard fields without bson tag
		if tag == "" {
			continue
		}
		// disregard secondary tag fragment
		tagStripped := strings.Split(tag, ",")
		if len(tagStripped) > 1 {
			tag = tagStripped[0]
			if tagStripped[0] == "_id" {
				continue
			}
		}
		if !utils.IsZeroType(val) {
			update := bson.E{Key: tag, Value: val.Interface()}
			updates = append(updates, update)
		}
	}
	updateFilter := bson.D{{Key: "$set", Value: updates}}

	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return mdb.Collection(opts.Table).UpdateOne(context.Background(), bson.M{
		"_id": _id,
	}, updateFilter)
}

func (m *MongoDriver) Delete(params any, opts DatabaseClientOptions) (count int, err error) {
	mdb := m.Client.Database(opts.Database)
	res, err := mdb.Collection(opts.Table).DeleteOne(context.Background(), params)
	return int(res.DeletedCount), err
}
