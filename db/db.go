package db

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
