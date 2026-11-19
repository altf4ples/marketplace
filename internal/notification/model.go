package notification

import (
	"github.com/google/uuid"
	"time"
)

type Notification struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	OrderID   *uuid.UUID `json:"order_id,omitempty"`
	Message   string     `json:"message"`
	CreatedAt time.Time  `json:"created_at"`
}
