package users

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	PhoneNumber *string   `json:"phone_number,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
}

func NewUser(name string, phone *string) (User, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 50 {
		return User{}, ErrNameRequired
	}
	if phone != nil {
		p := strings.TrimSpace(*phone)
		if p == "" {
			phone = nil
		} else {
			if len(p) > 15 {
				return User{}, ErrPhoneInvalid
			}
			phone = &p
		}
	}
	return User{ID: uuid.New(), Name: name, PhoneNumber: phone, CreatedAt: time.Now().UTC()}, nil
}
