package catalog

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"testing"
)

func TestNewProductValidation(t *testing.T) {
	if _, err := NewProduct(uuid.New(), "", nil, decimal.NewFromInt(1), nil); err != ErrNameRequired {
		t.Fatalf("expected name error, got %v", err)
	}
	if _, err := NewProduct(uuid.New(), "valid", nil, decimal.NewFromFloat(-1), nil); err != ErrPriceInvalid {
		t.Fatalf("expected price error, got %v", err)
	}
	if _, err := NewProduct(uuid.Nil, "valid", nil, decimal.NewFromInt(1), nil); err != ErrStoreIDInvalid {
		t.Fatalf("expected store id error, got %v", err)
	}
	if _, err := NewProduct(uuid.New(), "valid", nil, decimal.RequireFromString("1.001"), nil); err != ErrPriceInvalid {
		t.Fatalf("expected scale error, got %v", err)
	}
}
