package catalog

import "errors"

var (
	ErrNameRequired       = errors.New("name is required or exceeds the allowed length")
	ErrDescriptionInvalid = errors.New("description exceeds the allowed length")
	ErrPriceInvalid       = errors.New("price must be non-negative and have at most two decimal places")
	ErrStoreIDInvalid     = errors.New("store id is invalid")
	ErrImageURLInvalid    = errors.New("image_url must be an absolute http or https URL")
	ErrStoreNotFound      = errors.New("store not found")
	ErrProductNotFound    = errors.New("product not found")
	ErrConflict           = errors.New("resource already exists")
)
