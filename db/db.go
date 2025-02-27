package db

type DatabaseClientOptions struct {
	Database string
	// interchangeable with collection
	Table string
}

type DatabaseClient interface {
	Get(params any, opts DatabaseClientOptions, dest any) error
	GetOne(params any, opts DatabaseClientOptions) (any, error)
	CreateOne(document any, opts DatabaseClientOptions) (string, error)
	UpdateOne(id string, document any, opts DatabaseClientOptions) (any, error)
	Delete(params any, opts DatabaseClientOptions) (count int, err error)
}

//////////////////////////
// return types
//////////////////////////

type InsertedIDResponse struct {
	InsertedID string `json:"inserted_id"`
}

type MockDatabaseClient[T any] struct {
	data map[string]T
}

func NewMockDatabaseClient[T any]() *MockDatabaseClient[T] {
	return &MockDatabaseClient[T]{
		data: make(map[string]T),
	}
}

func (m *MockDatabaseClient[T]) AddData(key string, data T) {
	m.data[key] = data
}

func (m *MockDatabaseClient[T]) Get(params any, opts DatabaseClientOptions, dest any) error {
	return nil
}

func (m *MockDatabaseClient[T]) GetOne(params any, opts DatabaseClientOptions) (T, error) {
	return *new(T), nil
}

func (m *MockDatabaseClient[T]) CreateOne(document any, opts DatabaseClientOptions) (string, error) {
	// data, ok := document.(map[string]any)
	return "", nil
}

func (m *MockDatabaseClient[T]) UpdateOne(id string, document any, opts DatabaseClientOptions) (any, error) {
	return nil, nil
}

func (m *MockDatabaseClient[T]) Delete(params any, opts DatabaseClientOptions) (count int, err error) {
	return 0, nil
}
