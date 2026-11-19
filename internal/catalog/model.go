package catalog

import (
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Store struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type Product struct {
	ID          uuid.UUID       `json:"id"`
	StoreID     uuid.UUID       `json:"store_id"`
	StoreName   string          `json:"-"`
	Name        string          `json:"name"`
	Description *string         `json:"description,omitempty"`
	Price       decimal.Decimal `json:"price"`
	ImageURL    *string         `json:"image_url,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

func normalizeOptional(p *string) *string {
	if p == nil {
		return nil
	}
	s := strings.TrimSpace(*p)
	if s == "" {
		return nil
	}
	return &s
}

func NewStore(name string, description *string) (Store, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 50 {
		return Store{}, ErrNameRequired
	}

	description = normalizeOptional(description)
	if description != nil && len(*description) > 100 {
		return Store{}, ErrDescriptionInvalid
	}

	return Store{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		CreatedAt:   time.Now().UTC(),
	}, nil
}

func NewProduct(storeID uuid.UUID, name string, description *string, price decimal.Decimal, imageURL *string) (Product, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 200 {
		return Product{}, ErrNameRequired
	}
	if storeID == uuid.Nil {
		return Product{}, ErrStoreIDInvalid
	}
	if !price.GreaterThanOrEqual(decimal.Zero) || !price.Equal(price.Round(2)) {
		return Product{}, ErrPriceInvalid
	}
	description = normalizeOptional(description)
	if description != nil && len(*description) > 4000 {
		return Product{}, ErrDescriptionInvalid
	}
	imageURL = normalizeOptional(imageURL)
	if imageURL != nil {
		parsed, err := url.ParseRequestURI(*imageURL)
		if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return Product{}, ErrImageURLInvalid
		}
	}
	return Product{
		ID:          uuid.New(),
		StoreID:     storeID,
		Name:        name,
		Description: description,
		Price:       price,
		ImageURL:    imageURL,
		CreatedAt:   time.Now().UTC(),
	}, nil
}
