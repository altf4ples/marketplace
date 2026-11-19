package catalog

import (
	"time"

	"github.com/shopspring/decimal"
)

type StoreResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateStoreRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type ProductResponse struct {
	ID          string          `json:"id"`
	StoreID     string          `json:"store_id"`
	Name        string          `json:"name"`
	Description *string         `json:"description,omitempty"`
	Price       decimal.Decimal `json:"price"`
	ImageURL    *string         `json:"image_url,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

type CreateProductRequest struct {
	Name        string           `json:"name"`
	Description *string          `json:"description"`
	Price       *decimal.Decimal `json:"price"`
	ImageURL    *string          `json:"image_url,omitempty"`
}

type ProductListResponse struct {
	StoreID  string            `json:"store_id"`
	Products []ProductResponse `json:"products"`
}
