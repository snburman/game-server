package assets

type AssetType string

type Image struct {
	Name   string      `json:"name"`
	Path   string      `json:"path"`
	Width  int         `json:"width"`
	Height int         `json:"height"`
	Frames []FrameSpec `json:"frames"`
	Data   []byte      `json:"data"`
}

type FrameSpec struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}
