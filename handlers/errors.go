package handlers

const (
	ErrMissingParams ServerError = "missing_params"

	// Authentication
	//
	// Stytch Errors
	//
	// https://stytch.com/docs/api/password-authenticate#:~:text=will%20return%20a-,reset_password,-error%20even%20if
	// user must choose different email or login
	ErrDuplicateEmail AuthenticationError = "duplicate_email"
	ErrInvalidEmail   AuthenticationError = "invalid_email"
	// user must reset password
	ErrResetPassword    AuthenticationError = "reset_password"
	ErrBreachedPassword AuthenticationError = "breached_password"
	// user must choose different password
	ErrWeakPassword AuthenticationError = "weak_password"
	// Login Errors
	//
	ErrInvalidCredentials AuthenticationError = "invalid_credentials"
	// Profile Errors
	ErrCreatingProfile AuthenticationError = "error_creating_profile"
	ErrProfileExists   AuthenticationError = "user_exists"
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
