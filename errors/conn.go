package errors

type ConnectionError = ServerError

const (
	ErrConnectionExists   ConnectionError = "connection_exists"
	ErrConnectionNotFound ConnectionError = "connection_not_found"
)
