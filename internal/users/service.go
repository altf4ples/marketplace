package users

import (
	"context"
	"github.com/google/uuid"
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }
func (s *Service) CreateUser(ctx context.Context, name string, phone *string) (User, error) {
	u, err := NewUser(name, phone)
	if err != nil {
		return User{}, err
	}
	if err = s.repo.Create(ctx, u); err != nil {
		return User{}, err
	}
	return u, nil
}
func (s *Service) Exists(ctx context.Context, id uuid.UUID) error { return s.repo.Exists(ctx, id) }
func (s *Service) Get(ctx context.Context, id uuid.UUID) (User, error) {
	return s.repo.GetByID(ctx, id)
}
