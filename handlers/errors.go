package handlers

const (
	ErrMissingParams ServerError = "missing_params"
	// Login Errors
	//
	ErrInvalidCredentials AuthenticationError = "invalid_credentials"
	ErrWeakPassword       AuthenticationError = "weak_password"
	// User Errors
	ErrCreatingUser AuthenticationError = "error_creating_user"
	ErrUserExists   AuthenticationError = "user_exists"
	ErrUserBanned   AuthenticationError = "user_banned"
)

type ServerError string
type AuthenticationError = ServerError

func (e ServerError) Error() string {
	return string(e)
}

func (e ServerError) JSON() map[string]string {
	return map[string]string{
		"error": e.Error(),
	}
}
