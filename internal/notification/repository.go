package notification

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	ListByUser(context.Context, uuid.UUID, int) ([]Notification, error)
}
type PostgresRepo struct{ pool *pgxpool.Pool }

func NewPostgresRepo(pool *pgxpool.Pool) *PostgresRepo { return &PostgresRepo{pool: pool} }
func (r *PostgresRepo) ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]Notification, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,user_id,order_id,message,created_at FROM store.notifications WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Notification, 0)
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.OrderID, &n.Message, &n.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// CreateTx persists a notification in the caller's transaction, so order state
// and the corresponding notification commit or roll back together.
func (r *PostgresRepo) CreateTx(ctx context.Context, tx pgx.Tx, userID, orderID uuid.UUID, message string) error {
	_, err := tx.Exec(ctx, `INSERT INTO store.notifications(user_id,order_id,message) VALUES($1,$2,$3)`, userID, orderID, message)
	return err
}
