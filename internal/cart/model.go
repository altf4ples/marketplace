package cart

import "github.com/google/uuid"

type CartItem struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int64     `json:"quantity"`
}

type Cart struct {
	UserID uuid.UUID  `json:"user_id"`
	Items  []CartItem `json:"items"`
}
