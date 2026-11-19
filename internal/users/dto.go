package users

type UserResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	PhoneNumber *string `json:"phone_number,omitempty"`
}

type CreateUserRequest struct {
	Name        string  `json:"name"`
	PhoneNumber *string `json:"phone_number"`
}
