package catalog

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateStore(ctx context.Context, u Store) error
	GetByStoreID(ctx context.Context, id string) (Store, error)
	CreateProduct(ctx context.Context, u Product) error
	GetByProductID(ctx context.Context, id string) (Product, error)
	GetByStoreProducts(ctx context.Context, storeID string) ([]Product, error)
}

type PostgresRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresRepo(pool *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{pool: pool}
}

func mapWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrConflict
	}
	return err
}

func (r *PostgresRepo) CreateStore(ctx context.Context, u Store) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO store.stores (id, name, description, created_at) VALUES ($1, $2, $3, $4)`,
		u.ID, u.Name, u.Description, u.CreatedAt,
	)
	return mapWriteError(err)
}

func (r *PostgresRepo) GetByStoreID(ctx context.Context, id string) (Store, error) {
	var u Store
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, description, created_at FROM store.stores WHERE id = $1`, id,
	).Scan(&u.ID, &u.Name, &u.Description, &u.CreatedAt)
	return u, err
}

func (r *PostgresRepo) GetByStoreProducts(ctx context.Context, storeID string) ([]Product, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, store_id, name, description, price, image_url, created_at
         FROM store.products
         WHERE store_id = $1
         ORDER BY created_at`,
		storeID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]Product, 0)
	for rows.Next() {
		var p Product
		if err := rows.Scan(
			&p.ID,
			&p.StoreID,
			&p.Name,
			&p.Description,
			&p.Price,
			&p.ImageURL,
			&p.CreatedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func (r *PostgresRepo) CreateProduct(ctx context.Context, u Product) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO store.products (id, store_id, name, description, price, image_url) VALUES ($1, $2, $3, $4, $5, $6)`,
		u.ID, u.StoreID, u.Name, u.Description, u.Price, u.ImageURL,
	)
	return mapWriteError(err)
}

func (r *PostgresRepo) GetByProductID(ctx context.Context, id string) (Product, error) {
	var u Product
	err := r.pool.QueryRow(ctx,
		`SELECT p.id, p.store_id, p.name, p.description, p.price, p.image_url, p.created_at, s.name
         FROM store.products p JOIN store.stores s ON s.id=p.store_id WHERE p.id = $1`, id,
	).Scan(&u.ID, &u.StoreID, &u.Name, &u.Description, &u.Price, &u.ImageURL, &u.CreatedAt, &u.StoreName)
	return u, err
}
