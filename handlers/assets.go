package handlers

import (
	"context"
	"encoding/json"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/snburman/game_server/assets"
	"github.com/snburman/game_server/db"
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

func HandleCreatePlayerAsset(c echo.Context) error {
	asset := new(PlayerAsset[string])
	if err := c.Bind(&asset); err != nil {
		return err
	}
	byteAsset := new(PlayerAsset[[]byte])
	byteAsset.Name = asset.Name
	byteAsset.Width = asset.Width
	byteAsset.Height = asset.Height
	byteAsset.Data = []byte(asset.Data)

	_, err := db.MongoDB.Client.Database(db.GameDatabase).Collection(db.PlayerImagesCollection).
		InsertOne(context.Background(), byteAsset)
	if err != nil {
		log.Println("error in inserting image", err)
		return err
	}
	return c.JSON(200, asset)
}

func HandleGetPlayerAssets(c echo.Context) error {
	res, err := db.MongoDB.Client.Database(db.GameDatabase).Collection(db.PlayerImagesCollection).
		Find(context.Background(), bson.M{})
	if err != nil {
		return err
	}
	var imgs []PlayerAsset[[]byte]
	err = res.All(context.Background(), &imgs)
	if err != nil {
		log.Println(res)
		log.Println("error in fetching images", err)
		return err
	}

	var decodedImgs = []PlayerAsset[PixelData]{}
	for _, img := range imgs {
		_img := new(PlayerAsset[PixelData])
		// decode the json string
		err := json.Unmarshal(img.Data, &_img.Data)
		if err != nil {
			log.Println("error in decoding image", err)
			return err
		}
		_img.Name = img.Name
		_img.Width = img.Width
		_img.Height = img.Height
		decodedImgs = append(decodedImgs, *_img)
	}

	return c.JSON(200, decodedImgs)
}

type Pixel struct {
	X     int    `json:"x"`
	Y     int    `json:"y"`
	R     int    `json:"r"`
	G     int    `json:"g"`
	B     int    `json:"b"`
	A     int    `json:"a"`
	Color string `json:"color"`
}

type PixelData = [][]Pixel

type PlayerAsset[T any] struct {
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Name   string `json:"name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Data   T      `json:"data"`
}
