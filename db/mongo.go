package db

import (
	"context"
	"errors"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/snburman/game_server/config"
	"github.com/snburman/game_server/utils"
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

func (m *MongoDriver) CreateOne(document any, opts DatabaseClientOptions) (insertedID string, err error) {
	mdb := MongoDB.Client.Database(opts.Database)
	res, err := mdb.Collection(opts.Table).InsertOne(context.Background(), document)
	id := res.InsertedID.(primitive.ObjectID)
	return id.Hex(), err
}

func (m *MongoDriver) UpdateOne(id string, document any, opts DatabaseClientOptions) (any, error) {
	mdb := MongoDB.Client.Database(opts.Database)
	var updates bson.D

	typeData := reflect.TypeOf(document)
	values := reflect.ValueOf(document)

	// https://joshua-etim.medium.com/how-i-update-documents-in-mongodb-with-golang-94485dbe54f7
	for i := 1; i < typeData.NumField(); i++ {
		field := typeData.Field(i)
		val := values.Field(i)
		tag := field.Tag.Get("bson")
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
	mdb := MongoDB.Client.Database(opts.Database)
	res, err := mdb.Collection(opts.Table).DeleteOne(context.Background(), params)
	return int(res.DeletedCount), err
}

//////////////////////////////////
// Mock client implementation
//////////////////////////////////

type MockMongoClient struct {
	Fixtures map[string]any
}

func NewMockMongoClient(fixtures map[string]any) *MockMongoClient {
	return &MockMongoClient{
		Fixtures: fixtures,
	}
}

func (m *MockMongoClient) Get(params any, opts DatabaseClientOptions, dest *[]any) error {
	return nil
}

func (m *MockMongoClient) GetOne(params any, _ DatabaseClientOptions) (any, error) {
	b, err := bson.Marshal(params)
	if err != nil {
		return nil, err
	}
	res, ok := m.Fixtures[string(b)]
	if !ok {
		return nil, errors.New("document not found")
	}
	return res, nil
}

func (m *MockMongoClient) CreateOne(document any, _ DatabaseClientOptions) (string, error) {
	b, err := bson.Marshal(document)
	if err != nil {
		return "", err
	}
	m.Fixtures[string(b)] = document
	return primitive.NewObjectID().Hex(), nil
}

func (m *MockMongoClient) UpdateOne(id string, document any, opts DatabaseClientOptions) (any, error) {
	return nil, nil
}

func (m *MockMongoClient) Delete(params any, opts DatabaseClientOptions) (count int, err error) {
	b, err := bson.Marshal(params)
	if err != nil {
		return 0, err
	}
	initial := len(m.Fixtures)
	_, ok := m.Fixtures[string(b)]
	if !ok {
		return 0, errors.New("document not found")
	}
	delete(m.Fixtures, string(b))
	after := len(m.Fixtures)
	return initial - after, nil
}
