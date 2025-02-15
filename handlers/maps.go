package handlers

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/snburman/game-server/db"
	"github.com/snburman/game-server/errors"
	"github.com/snburman/game-server/middleware"
)

// HandleGetAllMaps retrieves all maps
func HandleGetAllMaps(c echo.Context) error {
	maps, err := db.GetAllMaps(db.MongoDB)
	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			errors.ServerError(err.Error()).JSON(),
		)
	}
	return c.JSON(http.StatusOK, maps)
}

// @QueryParam id
//
// @QueryParam userID
//
// HandleGetMapByID retrieves a map by ID and appends player character by userID
func HandleGetMapByID(c echo.Context) error {
	id := c.QueryParam("id")
	userID := c.QueryParam("userID")
	if id == "" || userID == "" {
		return c.JSON(
			http.StatusBadRequest,
			errors.ErrMissingParams.JSON(),
		)
	}

	_map, err := db.GetMapByID(db.MongoDB, id)
	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			errors.ServerError(err.Error()).JSON(),
		)
	}

	_map, err = db.AppendMapPlayerCharacter(db.MongoDB, userID, _map)
	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			errors.ServerError(err.Error()).JSON(),
		)
	}

	return c.JSON(http.StatusOK, _map)
}

// @QueryParam []string
//
// HandleGetAllMapsByIDs retrieves all maps by IDs
func HandleGetAllMapsByIDs(c echo.Context) error {
	ids := c.QueryParams()["ids"]
	if len(ids) == 0 {
		return c.JSON(
			http.StatusBadRequest,
			errors.ErrMissingParams.JSON(),
		)
	}

	maps, err := db.GetMapsByIDs(db.MongoDB, ids)
	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			errors.ServerError(err.Error()).JSON(),
		)
	}
	return c.JSON(http.StatusOK, maps)
}

// @Param userID
//
// HandleGetPrimaryMap retrieves the primary map by userID and appends player character
func HandleGetPlayerPrimaryMap(c echo.Context) error {
	userID := c.Param("userID")
	if userID == "" {
		return c.JSON(
			http.StatusBadRequest,
			errors.ErrMissingParams.JSON(),
		)
	}

	_map, err := db.GetPrimaryMapByUserID(db.MongoDB, userID)
	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			errors.ServerError(err.Error()).JSON(),
		)
	}

	_map, err = db.AppendMapPlayerCharacter(db.MongoDB, userID, _map)
	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			errors.ServerError(err.Error()).JSON(),
		)
	}
	return c.JSON(http.StatusOK, _map)
}

func HandleGetPlayerMaps(c echo.Context) error {
	claims, ok := c.(middleware.JWTContext)
	if !ok {
		return c.JSON(
			http.StatusUnauthorized,
			errors.ServerError(errors.ErrInvalidJWT).JSON())
	}
	maps, err := db.GetMapsByUserID(db.MongoDB, claims.UserID)
	if err != nil {
		return c.JSON(http.StatusNotFound, errors.ErrMapNotFound)
	}

	return c.JSON(http.StatusOK, maps)
}

// @Body Map[string]
//
// HandleCreateMap creates a new map
func HandleCreateMap(c echo.Context) error {
	claims, ok := c.(middleware.JWTContext)
	if !ok {
		return c.JSON(
			http.StatusUnauthorized,
			errors.ServerError(errors.ErrInvalidJWT).JSON())
	}
	_map := *new(db.Map[string])
	if err := c.Bind(&_map); err != nil {
		return c.JSON(
			http.StatusNotAcceptable,
			errors.ErrBindingPayload.JSON(),
		)
	}

	if claims.UserID != _map.UserID {
		return c.JSON(
			http.StatusUnauthorized,
			errors.ErrInvalidJWT.JSON(),
		)
	}

	insertedId, err := db.CreateMap(db.MongoDB, _map)
	if err != nil {
		log.Println(err)
		return c.JSON(
			http.StatusNotAcceptable,
			errors.ServerError(err.Error()).JSON(),
		)
	}
	return c.JSON(http.StatusAccepted, db.InsertedIDResponse{
		InsertedID: insertedId.Hex(),
	})
}

// @Body Map[string]
//
// HandleUpdateMap updates an existing map
func HandleUpdateMap(c echo.Context) error {
	claims, ok := c.(middleware.JWTContext)
	if !ok {
		return c.JSON(
			http.StatusUnauthorized,
			errors.ServerError(errors.ErrInvalidJWT).JSON())
	}

	_map := *new(db.Map[string])
	if err := c.Bind(&_map); err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			errors.ErrBindingPayload.JSON(),
		)
	}

	if claims.UserID != _map.UserID {
		return c.JSON(
			http.StatusUnauthorized,
			errors.ServerError(errors.ErrInvalidJWT).JSON())
	}

	existingMap, err := db.GetMapByNameUserID(db.MongoDB, _map.Name, _map.UserID)
	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			errors.ServerError(err.Error()).JSON(),
		)
	}

	_map.ID = existingMap.ID
	err = db.UpdateMap(db.MongoDB, _map)
	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			errors.ServerError(err.Error()).JSON(),
		)
	}

	return c.NoContent(http.StatusAccepted)
}

// @Param id
//
// HandleDeleteMap deletes a map by ID
func HandleDeleteMap(c echo.Context) error {
	claims, ok := c.(middleware.JWTContext)
	if !ok {
		return c.JSON(
			http.StatusUnauthorized,
			errors.ServerError(errors.ErrInvalidJWT).JSON())
	}

	id := c.Param("id")
	if id == "" {
		return c.JSON(
			http.StatusBadRequest,
			errors.ErrMissingParams.JSON())
	}

	_map, err := db.GetMapByID(db.MongoDB, id)
	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			errors.ServerError(err.Error()).JSON(),
		)
	}

	if claims.UserID != _map.UserID {
		return c.JSON(
			http.StatusUnauthorized,
			errors.ServerError(errors.ErrInvalidJWT).JSON())
	}

	err = db.DeleteMap(db.MongoDB, id)
	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			errors.ServerError(err.Error()).JSON(),
		)
	}

	return c.NoContent(http.StatusAccepted)
}
