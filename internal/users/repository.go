package users

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(context.Context, User) error
	GetByID(context.Context, uuid.UUID) (User, error)
	Exists(context.Context, uuid.UUID) error
}
type PostgresRepo struct{ pool *pgxpool.Pool }

func NewPostgresRepo(pool *pgxpool.Pool) *PostgresRepo { return &PostgresRepo{pool: pool} }

func (r *PostgresRepo) Create(ctx context.Context, u User) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO store.users (id, username, phone_number) VALUES ($1,$2,$3)`, u.ID, u.Name, u.PhoneNumber)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrConflict
	}
	return err
}
func (r *PostgresRepo) GetByID(ctx context.Context, id uuid.UUID) (User, error) {
	var u User
	err := r.pool.QueryRow(ctx, `SELECT id, username, phone_number FROM store.users WHERE id=$1`, id).Scan(&u.ID, &u.Name, &u.PhoneNumber)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	return u, err
}
func (r *PostgresRepo) Exists(ctx context.Context, id uuid.UUID) error {
	var exists bool
	if err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM store.users WHERE id=$1)`, id).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}
	return nil
}
