package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/snburman/game_server/assets"
	"github.com/snburman/game_server/db"
	"github.com/snburman/game_server/errors"
	"github.com/snburman/game_server/middleware"
	"go.mongodb.org/mongo-driver/bson"
)

func HandleGetAssets(c echo.Context) error {
	// get images from db
	res, err := db.MongoDB.Client.Database(db.GameDatabase).Collection(db.ImagesCollection).
		Find(context.Background(), bson.M{})
	if err != nil {
		return err
	}
	var imgs []assets.Image
	err = res.All(context.Background(), &imgs)
	if err != nil {
		log.Println(res)
		log.Println("error in fetching images", err)
		return err
	}
	return c.JSON(200, imgs)
}

// HandlePlayerGetAssets returns player assets by UserID in JWT claims
func HandleGetPlayerAssets(c echo.Context) error {
	// get user id from claims
	claims, ok := c.(middleware.JWTContext)
	if !ok {
		return c.JSON(
			http.StatusUnauthorized,
			errors.ServerError(errors.ErrInvalidJWT).JSON())
	}

	assets, err := db.GetPlayerAssetsByUserID(db.MongoDB, claims.UserID)
	if err != nil {
		log.Println(err)
		return c.JSON(
			http.StatusNotFound,
			errors.ServerError("error_getting_assets").JSON())
	}

	return c.JSON(200, assets)
}

func HandleCreatePlayerAsset(c echo.Context) error {
	// get user id from claims
	claims, ok := c.(middleware.JWTContext)
	if !ok {
		return c.JSON(
			http.StatusUnauthorized,
			errors.ServerError(errors.ErrInvalidJWT).JSON())
	}
	// get asset from req body
	asset := *new(db.PlayerAsset[string])
	if err := c.Bind(&asset); err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			errors.ErrBindingPayload.JSON(),
		)
	}
	// reject if claims and asset userID do not match
	if claims.UserID != asset.UserID {
		return c.JSON(
			http.StatusUnauthorized,
			errors.ServerError(errors.ErrInvalidJWT).JSON())
	}
	// create asset
	id, err := db.CreatePlayerAsset(db.MongoDB, asset)
	if err != nil {
		if err == errors.ErrImageExists {
			return c.JSON(http.StatusNotAcceptable,
				errors.ErrImageExists.JSON())
		}
		return c.JSON(http.StatusInternalServerError,
			errors.ServerError(err.Error()).JSON())
	}

	return c.JSON(http.StatusAccepted, db.InsertedIDResponse{
		InsertedID: id.Hex(),
	})
}

func HandleUpdatePlayerAsset(c echo.Context) error {
	// get user id from claims
	claims, ok := c.(middleware.JWTContext)
	if !ok {
		return c.JSON(
			http.StatusUnauthorized,
			errors.ServerError(errors.ErrInvalidJWT).JSON(),
		)
	}
	// get asset from req body
	asset := *new(db.PlayerAsset[string])
	if err := c.Bind(&asset); err != nil {
		return err
	}
	// reject if claims and asset userID do not match
	if claims.UserID != asset.UserID {
		return c.JSON(
			http.StatusUnauthorized,
			errors.ServerError(errors.ErrInvalidJWT).JSON(),
		)
	}

	// get asset by userID and name
	existingAsset, err := db.GetPlayerAssetByNameUserID(db.MongoDB, asset.Name, asset.UserID)
	if err != nil {
		switch err {
		case errors.ErrImageNotFound:
			return c.JSON(
				http.StatusNotFound,
				errors.AssetError(err.Error()).JSON(),
			)
		}
	}
	// assign existing ID
	asset.ID = existingAsset.ID

	// update asset
	if err = db.UpdatePlayerAsset(db.MongoDB, asset); err != nil {
		log.Println(err)
		return c.JSON(
			http.StatusInternalServerError,
			errors.ServerError(err.Error()).JSON(),
		)
	}
	return c.NoContent(http.StatusAccepted)
}

func HandleDeletePlayerAsset(c echo.Context) error {
	imageID := c.QueryParam("id")
	if imageID == "" {
		return c.JSON(http.StatusBadRequest, errors.ErrMissingParams.JSON())
	}
	count, err := db.DeletePlayerAsset(db.MongoDB, imageID)
	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			errors.ServerError(err.Error()).JSON())
	}

	return c.JSON(http.StatusAccepted, struct {
		Deleted int `json:"deleted"`
	}{
		Deleted: count,
	})
}
