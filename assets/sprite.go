package assets

import (
	"log"

	"github.com/hajimehoshi/ebiten"
)

type Image struct {
	Name          string      `json:"name"`
	Path          string      `json:"path"`
	Width         int         `json:"width"`
	Height        int         `json:"height"`
	Frames        []FrameSpec `json:"frames"`
	Data          []byte      `json:"data"`
	*ebiten.Image `json:"-"`
}

func GenerateSprites(assets *Assets) {
	count := len(assets.Images.Sprites)
	log.Printf("Generating %d characters", count)
	for k, sprite := range assets.Images.Sprites {
		log.Printf("generating sprite: %s; source: %s", sprite.Name, sprite.Path)

		// Load the sprite data
		png, err := NewAsset(PNG, sprite.Path)
		if err != nil {
			log.Printf("error loading sprite: %s", err)
			continue
		}
		s := assets.Images.Sprites[k]
		s.Data = png

		// Generate the sprite image
		img, err := imageFromBytes(png)
		if err != nil {
			log.Printf("error decoding sprite: %s", err)
			continue
		}
		s.Image = img

		assets.Images.Sprites[k] = s
		log.Printf("generated sprite size: %d", len(s.Data))
	}
}
