package order

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"time"
)

type Status string

const (
	StatusCreated    Status = "created"
	StatusAssembling Status = "assembling"
	StatusShipped    Status = "shipped"
	StatusDelivered  Status = "delivered"
	StatusCancelled  Status = "cancelled"
)

type Event string

const (
	EventCreated  Event = "created"
	EventAssemble Event = "assemble"
	EventShip     Event = "ship"
	EventDeliver  Event = "deliver"
	EventCancel   Event = "cancel"
)

type Product struct {
	ID        uuid.UUID
	StoreID   uuid.UUID
	Name      string
	StoreName string
	Price     decimal.Decimal
}
type OrderItem struct {
	ID          uuid.UUID       `json:"id"`
	ProductID   uuid.UUID       `json:"product_id"`
	StoreID     uuid.UUID       `json:"store_id"`
	ProductName string          `json:"product_name"`
	StoreName   string          `json:"store_name"`
	Quantity    int64           `json:"quantity"`
	Price       decimal.Decimal `json:"price"`
	LineTotal   decimal.Decimal `json:"line_total"`
}
type Order struct {
	ID         uuid.UUID       `json:"id"`
	UserID     uuid.UUID       `json:"user_id"`
	Status     Status          `json:"status"`
	TotalPrice decimal.Decimal `json:"total_price"`
	Items      []OrderItem     `json:"items"`
	History    []StateChange   `json:"history"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}
type StateChange struct {
	From      *Status   `json:"from"`
	To        Status    `json:"to"`
	Event     Event     `json:"event"`
	CreatedAt time.Time `json:"created_at"`
}
