package errors

type MapError = ServerError

const (
	ErrMapExists      MapError = "map_exists"
	ErrMapNotFound    MapError = "map_not_found"
	ErrCreatingMap    MapError = "error_creating_map"
	ErrMapWrongFormat MapError = "map_wrong_format"
)
