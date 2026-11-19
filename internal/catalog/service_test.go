package catalog

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type catalogRepoStub struct {
	store            Store
	product          Product
	products         []Product
	getStoreErr      error
	createStoreErr   error
	createProductErr error
}

func (r *catalogRepoStub) CreateStore(_ context.Context, s Store) error {
	r.store = s
	return r.createStoreErr
}
func (r *catalogRepoStub) GetByStoreID(context.Context, string) (Store, error) {
	return r.store, r.getStoreErr
}
func (r *catalogRepoStub) CreateProduct(_ context.Context, p Product) error {
	r.product = p
	return r.createProductErr
}
func (r *catalogRepoStub) GetByProductID(context.Context, string) (Product, error) {
	return r.product, nil
}
func (r *catalogRepoStub) GetByStoreProducts(context.Context, string) ([]Product, error) {
	return r.products, nil
}

func TestCreateStoreValidatesAndPersists(t *testing.T) {
	repo := &catalogRepoStub{}
	svc := NewService(repo)
	if _, err := svc.CreateStores(context.Background(), " ", nil); !errors.Is(err, ErrNameRequired) {
		t.Fatalf("got %v", err)
	}
	if repo.store.ID != uuid.Nil {
		t.Fatal("invalid store reached repository")
	}
	got, err := svc.CreateStores(context.Background(), " Shop ", nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Shop" || repo.store.ID != got.ID {
		t.Fatalf("unexpected persisted store: %#v", got)
	}
}

func TestCreateProductRequiresExistingStoreAndValidProduct(t *testing.T) {
	storeID := uuid.New()
	repo := &catalogRepoStub{getStoreErr: pgx.ErrNoRows}
	svc := NewService(repo)
	_, err := svc.CreateProducts(context.Background(), storeID, "Book", nil, decimal.NewFromInt(10), nil)
	if !errors.Is(err, ErrStoreNotFound) {
		t.Fatalf("got %v", err)
	}

	repo.getStoreErr = nil
	_, err = svc.CreateProducts(context.Background(), storeID, "Book", nil, decimal.RequireFromString("10.001"), nil)
	if !errors.Is(err, ErrPriceInvalid) {
		t.Fatalf("got %v", err)
	}
	if repo.product.ID != uuid.Nil {
		t.Fatal("invalid product reached repository")
	}

	got, err := svc.CreateProducts(context.Background(), storeID, " Book ", nil, decimal.RequireFromString("10.50"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Book" || got.StoreID != storeID || repo.product.ID != got.ID {
		t.Fatalf("unexpected product: %#v", got)
	}
}

func TestCreateProductPropagatesPersistenceError(t *testing.T) {
	want := errors.New("write failed")
	repo := &catalogRepoStub{store: Store{ID: uuid.New()}, createProductErr: want}
	svc := NewService(repo)
	_, err := svc.CreateProducts(context.Background(), repo.store.ID, "Book", nil, decimal.NewFromInt(2), nil)
	if !errors.Is(err, want) {
		t.Fatalf("got %v, want %v", err, want)
	}
}
