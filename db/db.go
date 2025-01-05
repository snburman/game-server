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
