package db

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson"
)

type DatabaseClientOptions struct {
	Database string
	Table    string
}

type DatabaseClient interface {
	Get(params any, opts DatabaseClientOptions, dest *[]any) error
	GetOne(params any, opts DatabaseClientOptions) (any, error)
	CreateOne(document any, opts DatabaseClientOptions) error
	UpdateOne(document any, opts DatabaseClientOptions) (any, error)
	Delete(params any, opts DatabaseClientOptions) (count int, err error)
}

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

func (m *MockMongoClient) GetOne(params any, opts DatabaseClientOptions) (any, error) {
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

func (m *MockMongoClient) CreateOne(document any, opts DatabaseClientOptions) error {
	b, err := bson.Marshal(document)
	if err != nil {
		return err
	}
	m.Fixtures[string(b)] = document
	return nil
}

func (m *MockMongoClient) UpdateOne(document any, opts DatabaseClientOptions) (any, error) {
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
