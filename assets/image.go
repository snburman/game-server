package assets

import (
	"bytes"
	"image"
	"image/png"
	"io"

	"github.com/hajimehoshi/ebiten"
)

const (
	PNG AssetType = "png"
)

func pngBytesFromFile(file io.Reader) ([]byte, error) {
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	err = png.Encode(buf, img)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func imageFromBytes(data []byte) (*ebiten.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	eImg, err := ebiten.NewImageFromImage(img, ebiten.FilterDefault)
	return eImg, err
}
