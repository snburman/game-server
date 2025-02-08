package assets

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
)

type AssetType string

type Asset struct {
	Path string
	Data []byte
}

func NewAsset(t AssetType, path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	switch t {
	case PNG:
		return pngBytesFromFile(file)
	default:
		return nil, errors.New("unsupported asset type")
	}
}

type Assets struct {
	Images Images `json:"images"`
}

func (a *Assets) Sprite(name string) Image {
	return a.Images.Sprites[name]
}

type Images struct {
	Sprites map[string]Image `json:"sprites"`
}

type FrameSpec struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

func GenerateAssets() {
	file, err := os.Open("assets/config.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	b, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	var assetConf Assets
	err = json.Unmarshal(b, &assetConf)
	if err != nil {
		panic(err)
	}

	writeAssetsData(assetConf)
}

// writeAssetsData writes the assets data to the data.assets.json file
func writeAssetsData(assetConf Assets) {
	log.Println("writing assets config")
	f, err := os.Create("assets/data.assets.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	err = enc.Encode(assetConf)
	if err != nil {
		panic(err)
	}
}
