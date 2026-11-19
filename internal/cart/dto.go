package cart

import "github.com/google/uuid"

type AddItemRequest struct {
	UserID    uuid.UUID `json:"user_id"`
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int64     `json:"quantity"`
}

type CartResponse struct {
	UserID uuid.UUID  `json:"user_id"`
	Items  []CartItem `json:"items"`
}

func toCartResponse(c Cart) CartResponse {
	if c.Items == nil {
		c.Items = []CartItem{}
	}
	return CartResponse{UserID: c.UserID, Items: c.Items}
}
