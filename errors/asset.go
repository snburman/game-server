package errors

type AssetError = ServerError

const (
	ErrImageExists      AssetError = "image_exists"
	ErrImageNotFound    AssetError = "image_not_found"
	ErrCreatingImage    AssetError = "error_creating_image"
	ErrImageWrongFormat AssetError = "image_wrong_format"
)
