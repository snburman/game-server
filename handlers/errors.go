package handlers

type Error string

func (e Error) Error() string {
	return string(e)
}

func (e Error) JSON() map[string]string {
	return map[string]string{
		"error": e.Error(),
	}
}

const (
	ErrInvalidRequest     Error = "invalid request"
	ErrInvalidCredentials Error = "invalid credentials"
	ErrInvalidSession     Error = "invalid session"
	ErrUserExists         Error = "user already exists"
	ErrCreatingSession    Error = "error creating session"
	ErrCreatingUser       Error = "error creating user"
	ErrFetchingImages     Error = "error fetching images"
	ErrLoggingIn          Error = "error logging in"
	ErrLoggingOut         Error = "error logging out"
	ErrRegistering        Error = "error registering"
)
