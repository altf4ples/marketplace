package order

import "errors"

var (
	ErrCartEmpty         = errors.New("cart is empty")
	ErrUserNotFound      = errors.New("user not found")
	ErrProductNotFound   = errors.New("product not found")
	ErrInvalidQuantity   = errors.New("invalid item quantity")
	ErrInvalidTransition = errors.New("event is not allowed for current order status")
	ErrOrderNotFound     = errors.New("order not found")
	ErrForbidden         = errors.New("order does not belong to user")
)

type CartRestoreError struct {
	Cause      error
	RestoreErr error
}

func (e *CartRestoreError) Error() string {
	return "checkout failed and cart restoration failed: " + e.Cause.Error() + "; " + e.RestoreErr.Error()
}
func (e *CartRestoreError) Unwrap() []error { return []error{e.Cause, e.RestoreErr} }
