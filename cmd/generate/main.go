package main

import (
	"context"

	"github.com/snburman/game_server/assets"
	"github.com/snburman/game_server/db"
)

func main() {
	assets.GenerateAssets()
	db.NewMongoDriver()

	assets := assets.Load()

	var imgs []interface{} = []interface{}{}
	for _, sprite := range assets.Images.Sprites {
		imgs = append(imgs, sprite)
	}
	res, err := db.MongoDB.Client.Database(db.GameDatabase).Collection(db.ImagesCollection).
		InsertMany(context.Background(), imgs)
	if err != nil {
		panic(err)
	}
	println(res.InsertedIDs)
}
