package cart

import "errors"

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrProductNotFound = errors.New("product not found")
	ErrCartNotFound    = errors.New("cart not found")
	ErrStoreNotFound   = errors.New("store not found")
	ErrQuantityInvalid = errors.New("quantity must be between 1 and 1000")
)
