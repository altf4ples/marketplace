package order

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type fakeCart struct {
	items    []CartItem
	restored []CartItem
	takeErr  error
}

func (f *fakeCart) Take(context.Context, uuid.UUID) ([]CartItem, error) { return f.items, f.takeErr }
func (f *fakeCart) Restore(_ context.Context, _ uuid.UUID, items []CartItem) error {
	f.restored = append(f.restored, items...)
	return nil
}

type fakeProducts struct {
	items map[uuid.UUID]Product
	err   error
}

func (f fakeProducts) GetProduct(_ context.Context, id uuid.UUID) (Product, error) {
	if f.err != nil {
		return Product{}, f.err
	}
	p, ok := f.items[id]
	if !ok {
		return Product{}, ErrProductNotFound
	}
	return p, nil
}

type fakeUsers struct{ err error }

func (f fakeUsers) Exists(context.Context, uuid.UUID) error { return f.err }

type fakeOrders struct {
	created Order
	err     error
}

func (f *fakeOrders) Create(_ context.Context, o Order) (Order, error) {
	f.created = o
	return o, f.err
}
func (f *fakeOrders) Get(context.Context, uuid.UUID, uuid.UUID) (Order, error) { return Order{}, nil }
func (f *fakeOrders) ListByUser(context.Context, uuid.UUID) ([]Order, error)   { return nil, nil }
func (f *fakeOrders) Transition(context.Context, uuid.UUID, uuid.UUID, Event) (Order, error) {
	return Order{}, nil
}

func TestCheckoutCreatesPriceSnapshot(t *testing.T) {
	userID, productID, storeID := uuid.New(), uuid.New(), uuid.New()
	cart := &fakeCart{items: []CartItem{{ProductID: productID, Quantity: 3}}}
	products := fakeProducts{items: map[uuid.UUID]Product{productID: {ID: productID, StoreID: storeID, Name: "Keyboard", StoreName: "Shop", Price: decimal.RequireFromString("12.50")}}}
	repo := &fakeOrders{}
	svc := NewService(repo, cart, products, fakeUsers{})
	got, err := svc.Checkout(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if got.TotalPrice.String() != "37.5" {
		t.Fatalf("total = %s", got.TotalPrice)
	}
	if len(got.Items) != 1 || got.Items[0].Price.String() != "12.5" || got.Items[0].ProductName != "Keyboard" {
		t.Fatalf("bad order snapshot: %#v", got.Items)
	}
	if len(cart.restored) != 0 {
		t.Fatal("cart should not be restored after successful checkout")
	}
}
func TestCheckoutRestoresCartWhenProductMissing(t *testing.T) {
	userID, productID := uuid.New(), uuid.New()
	cart := &fakeCart{items: []CartItem{{ProductID: productID, Quantity: 1}}}
	svc := NewService(&fakeOrders{}, cart, fakeProducts{items: map[uuid.UUID]Product{}}, fakeUsers{})
	_, err := svc.Checkout(context.Background(), userID)
	if !errors.Is(err, ErrProductNotFound) {
		t.Fatalf("got %v", err)
	}
	if len(cart.restored) != 1 || cart.restored[0].ProductID != productID {
		t.Fatalf("cart not restored: %#v", cart.restored)
	}
}
func TestCheckoutRestoresCartWhenDatabaseFails(t *testing.T) {
	userID, productID := uuid.New(), uuid.New()
	cart := &fakeCart{items: []CartItem{{ProductID: productID, Quantity: 1}}}
	repo := &fakeOrders{err: errors.New("database unavailable")}
	svc := NewService(repo, cart, fakeProducts{items: map[uuid.UUID]Product{productID: {ID: productID, StoreID: uuid.New(), Name: "Book", StoreName: "Shop", Price: decimal.NewFromInt(10)}}}, fakeUsers{})
	_, err := svc.Checkout(context.Background(), userID)
	if err == nil {
		t.Fatal("expected repository error")
	}
	if len(cart.restored) != 1 {
		t.Fatalf("cart not restored: %#v", cart.restored)
	}
}

func TestCheckoutRejectsUnknownUserBeforeTakingCart(t *testing.T) {
	cart := &fakeCart{items: []CartItem{{ProductID: uuid.New(), Quantity: 1}}}
	svc := NewService(&fakeOrders{}, cart, fakeProducts{}, fakeUsers{err: ErrUserNotFound})
	_, err := svc.Checkout(context.Background(), uuid.New())
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("got %v, want ErrUserNotFound", err)
	}
	if len(cart.restored) != 0 {
		t.Fatal("cart must not be taken or restored when user validation fails")
	}
}

func TestCheckoutRejectsEmptyCart(t *testing.T) {
	cart := &fakeCart{}
	svc := NewService(&fakeOrders{}, cart, fakeProducts{}, fakeUsers{})
	_, err := svc.Checkout(context.Background(), uuid.New())
	if !errors.Is(err, ErrCartEmpty) {
		t.Fatalf("got %v, want ErrCartEmpty", err)
	}
}

func TestCheckoutRestoresCartWhenQuantityIsInvalid(t *testing.T) {
	userID, productID := uuid.New(), uuid.New()
	cart := &fakeCart{items: []CartItem{{ProductID: productID, Quantity: 1001}}}
	svc := NewService(&fakeOrders{}, cart, fakeProducts{}, fakeUsers{})
	_, err := svc.Checkout(context.Background(), userID)
	if !errors.Is(err, ErrInvalidQuantity) {
		t.Fatalf("got %v, want ErrInvalidQuantity", err)
	}
	if len(cart.restored) != 1 || cart.restored[0].ProductID != productID {
		t.Fatalf("cart was not restored: %#v", cart.restored)
	}
}
