package errors

const (
	ErrMissingParams  ServerError = "missing_params"
	ErrInvalidJWT     ServerError = "invalid_jwt"
	ErrBindingPayload ServerError = "error_binding_payload"
	ErrServerError    ServerError = "server_error"
)

type ServerError string

func (e ServerError) Error() string {
	return string(e)
}

func (e ServerError) JSON() map[string]string {
	return map[string]string{
		"error": e.Error(),
	}
}
