package handlers

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/snburman/game_server/db"
	"github.com/snburman/game_server/errors"
	"github.com/snburman/game_server/middleware"
)

func HandleGetAllMaps(c echo.Context) error {
	return nil
}

func HandleGetMapByID(c echo.Context) error {
	return nil
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
			errors.ErrCreatingMap.JSON(),
		)
	}
	return c.JSON(http.StatusAccepted, db.InsertedIDResponse{
		InsertedID: insertedId.Hex(),
	})
}

func HandleUpdateMap(c echo.Context) error {
	return nil
}

func HandleDeleteMap(c echo.Context) error {
	return nil
}
