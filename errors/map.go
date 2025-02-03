package errors

type MapError = ServerError

const (
	ErrMapExists        MapError = "map_exists"
	ErrPrimaryMapExists MapError = "primary_map_exists"
	ErrMapNotFound      MapError = "map_not_found"
	ErrCreatingMap      MapError = "error_creating_map"
	ErrUpdatingMap      MapError = "error_updating_map"
	ErrMapWrongFormat   MapError = "map_wrong_format"
)
