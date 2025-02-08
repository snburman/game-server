package assets

import (
	"bytes"
	"image"
	"image/png"
	"io"
)

const (
	PNG AssetType = "png"
)

type Image struct {
	Name   string      `json:"name"`
	Path   string      `json:"path"`
	Width  int         `json:"width"`
	Height int         `json:"height"`
	Frames []FrameSpec `json:"frames"`
	Data   []byte      `json:"data"`
}

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
