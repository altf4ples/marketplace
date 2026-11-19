package users

import "errors"

var (
	ErrNameRequired = errors.New("name must contain 1..50 characters")
	ErrPhoneInvalid = errors.New("phone_number must contain at most 15 characters")
	ErrNotFound     = errors.New("user not found")
	ErrConflict     = errors.New("user already exists")
)
