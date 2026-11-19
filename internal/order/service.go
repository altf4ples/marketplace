package order

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CartItem struct {
	ProductID uuid.UUID
	Quantity  int64
}
type CartReader interface {
	Take(context.Context, uuid.UUID) ([]CartItem, error)
	Restore(context.Context, uuid.UUID, []CartItem) error
}
type ProductReader interface {
	GetProduct(context.Context, uuid.UUID) (Product, error)
}
type UserReader interface {
	Exists(context.Context, uuid.UUID) error
}
type Service struct {
	repo     Repository
	cart     CartReader
	products ProductReader
	users    UserReader
}

func NewService(repo Repository, cart CartReader, products ProductReader, users UserReader) *Service {
	return &Service{repo: repo, cart: cart, products: products, users: users}
}

func (s *Service) Checkout(ctx context.Context, userID uuid.UUID) (Order, error) {
	if userID == uuid.Nil {
		return Order{}, ErrUserNotFound
	}
	if err := s.users.Exists(ctx, userID); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return Order{}, ErrUserNotFound
		}
		return Order{}, fmt.Errorf("check user: %w", err)
	}
	cartItems, err := s.cart.Take(ctx, userID)
	if err != nil {
		return Order{}, fmt.Errorf("take cart: %w", err)
	}
	if len(cartItems) == 0 {
		return Order{}, ErrCartEmpty
	}
	restore := func(cause error) (Order, error) {
		restoreCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if restoreErr := s.cart.Restore(restoreCtx, userID, cartItems); restoreErr != nil {
			return Order{}, &CartRestoreError{Cause: cause, RestoreErr: fmt.Errorf("restore cart: %w", restoreErr)}
		}
		return Order{}, cause
	}
	orderItems := make([]OrderItem, 0, len(cartItems))
	total := decimal.Zero
	for _, ci := range cartItems {
		if ci.Quantity < 1 || ci.Quantity > 1000 {
			return restore(ErrInvalidQuantity)
		}
		p, err := s.products.GetProduct(ctx, ci.ProductID)
		if err != nil {
			if errors.Is(err, ErrProductNotFound) {
				return restore(ErrProductNotFound)
			}
			return restore(fmt.Errorf("load product: %w", err))
		}
		line := p.Price.Mul(decimal.NewFromInt(ci.Quantity))
		total = total.Add(line)
		orderItems = append(orderItems, OrderItem{ProductID: p.ID, StoreID: p.StoreID, ProductName: p.Name, StoreName: p.StoreName, Quantity: ci.Quantity, Price: p.Price, LineTotal: line})
	}
	o := Order{ID: uuid.New(), UserID: userID, Status: StatusCreated, TotalPrice: total, Items: orderItems}
	created, err := s.repo.Create(ctx, o)
	if err != nil {
		return restore(fmt.Errorf("create order: %w", err))
	}
	return created, nil
}
func (s *Service) Get(ctx context.Context, id, userID uuid.UUID) (Order, error) {
	return s.repo.Get(ctx, id, userID)
}
func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]Order, error) {
	if err := s.users.Exists(ctx, userID); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("check user: %w", err)
	}
	return s.repo.ListByUser(ctx, userID)
}
func (s *Service) ApplyEvent(ctx context.Context, id, userID uuid.UUID, event Event) (Order, error) {
	if id == uuid.Nil || userID == uuid.Nil {
		return Order{}, ErrOrderNotFound
	}
	return s.repo.Transition(ctx, id, userID, event)
}
