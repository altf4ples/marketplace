package order

import "github.com/google/uuid"

type CheckoutRequest struct {
	UserID uuid.UUID `json:"user_id"`
}
type EventRequest struct {
	UserID uuid.UUID `json:"user_id"`
	Event  Event     `json:"event"`
}
