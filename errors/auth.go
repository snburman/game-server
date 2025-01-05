package errors

const (
	// Login Errors
	//
	ErrInvalidCredentials AuthenticationError = "invalid_credentials"
	ErrWeakPassword       AuthenticationError = "weak_password"
	// User Errors
	ErrUserExists   AuthenticationError = "user_exists"
	ErrCreatingUser AuthenticationError = "error_creating_user"
	ErrUpdatingUser AuthenticationError = "error_updating_user"
	ErrUserBanned   AuthenticationError = "user_banned"
)

type AuthenticationError = ServerError
