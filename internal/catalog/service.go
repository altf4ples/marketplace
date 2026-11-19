package catalog

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateStores(ctx context.Context, name string, description *string) (Store, error) {
	u, err := NewStore(name, description)
	if err != nil {
		return Store{}, err
	}

	if err := s.repo.CreateStore(ctx, u); err != nil {
		return Store{}, err
	}
	return u, nil
}

func (s *Service) GetByStoreProducts(ctx context.Context, storeID string) ([]Product, error) {
	if _, err := s.repo.GetByStoreID(ctx, storeID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []Product{}, ErrStoreNotFound
		}
		return []Product{}, err
	}

	products, err := s.repo.GetByStoreProducts(ctx, storeID)
	if err != nil {
		return []Product{}, err
	}
	return products, nil
}

func (s *Service) CreateProducts(ctx context.Context, storeID uuid.UUID, name string, description *string, price decimal.Decimal, imageURL *string) (Product, error) {
	if _, err := s.repo.GetByStoreID(ctx, storeID.String()); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Product{}, ErrStoreNotFound
		}
		return Product{}, err
	}

	p, err := NewProduct(storeID, name, description, price, imageURL)
	if err != nil {
		return Product{}, err
	}
	if err := s.repo.CreateProduct(ctx, p); err != nil {
		return Product{}, err
	}
	return p, nil
}

// GetProduct returns a product for other application modules through a narrow read interface.
func (s *Service) GetProduct(ctx context.Context, id uuid.UUID) (Product, error) {
	p, err := s.repo.GetByProductID(ctx, id.String())
	if errors.Is(err, pgx.ErrNoRows) {
		return Product{}, ErrProductNotFound
	}
	return p, err
}
