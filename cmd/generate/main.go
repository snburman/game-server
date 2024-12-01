package main

import (
	"context"

	"github.com/snburman/magic_game_server/assets"
	"github.com/snburman/magic_game_server/db"
)

func main() {
	assets.GenerateAssets()
	db.NewMongoDriver()

	assets := assets.Load()

	var imgs []interface{} = []interface{}{}
	for _, sprite := range assets.Images.Sprites {
		imgs = append(imgs, sprite)
	}
	res, err := db.MongoDB.Client.Database("magic_game").Collection("images").
		InsertMany(context.Background(), imgs)
	if err != nil {
		panic(err)
	}
	println(res.InsertedIDs)
}
