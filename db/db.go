package db

type DatabaseClientOptions struct {
	Database string
	Table    string
}

type DatabaseClient interface {
	Get(params any, opts DatabaseClientOptions, dest *[]any) error
	GetOne(id string, opts DatabaseClientOptions, dest *any) error
	CreateOne(document any, opts DatabaseClientOptions) (any, error)
	UpdateOne(document any, opts DatabaseClientOptions) (any, error)
	Delete(params any, opts DatabaseClientOptions) (int, error)
}
