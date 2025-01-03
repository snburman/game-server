package handlers

const (
	ErrMissingParams ServerError = "missing_params"

	// user must choose different password
	ErrWeakPassword AuthenticationError = "weak_password"
	// Login Errors
	//
	ErrInvalidCredentials AuthenticationError = "invalid_credentials"
	// Profile Errors
	ErrCreatingProfile AuthenticationError = "error_creating_profile"
	ErrUserExists      AuthenticationError = "user_exists"
	ErrUserBanned      AuthenticationError = "user_banned"
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
