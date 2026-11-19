package cart

import (
	"context"
	"github.com/google/uuid"
)

type Service struct{ repo CartRepository }

func NewService(repo CartRepository) *Service { return &Service{repo: repo} }

func (s *Service) GetOrCreate(ctx context.Context, userID uuid.UUID) (Cart, error) {
	if userID == uuid.Nil {
		return Cart{}, ErrUserNotFound
	}
	return s.repo.GetOrCreate(ctx, userID)
}
func (s *Service) AddItem(ctx context.Context, userID, productID uuid.UUID, qty int64) (Cart, error) {
	if userID == uuid.Nil {
		return Cart{}, ErrUserNotFound
	}
	if productID == uuid.Nil {
		return Cart{}, ErrProductNotFound
	}
	if qty <= 0 || qty > 1000 {
		return Cart{}, ErrQuantityInvalid
	}
	return s.repo.AddItem(ctx, userID, productID, qty)
}
func (s *Service) RemoveItem(ctx context.Context, userID, productID uuid.UUID) (Cart, error) {
	if userID == uuid.Nil {
		return Cart{}, ErrUserNotFound
	}
	if productID == uuid.Nil {
		return Cart{}, ErrProductNotFound
	}
	return s.repo.RemoveItem(ctx, userID, productID)
}
func (s *Service) Get(ctx context.Context, userID uuid.UUID) (Cart, error) {
	if userID == uuid.Nil {
		return Cart{}, ErrUserNotFound
	}
	return s.repo.Get(ctx, userID)
}
func (s *Service) Clear(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return ErrUserNotFound
	}
	return s.repo.Clear(ctx, userID)
}
func (s *Service) Take(ctx context.Context, userID uuid.UUID) ([]CartItem, error) {
	if userID == uuid.Nil {
		return nil, ErrUserNotFound
	}
	return s.repo.Take(ctx, userID)
}

func (s *Service) Restore(ctx context.Context, userID uuid.UUID, items []CartItem) error {
	if userID == uuid.Nil {
		return ErrUserNotFound
	}
	return s.repo.Restore(ctx, userID, items)
}
