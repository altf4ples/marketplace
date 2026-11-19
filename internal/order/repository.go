package order

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type NotificationWriter interface {
	CreateTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, string) error
}

type Repository interface {
	Create(context.Context, Order) (Order, error)
	Get(context.Context, uuid.UUID, uuid.UUID) (Order, error)
	ListByUser(context.Context, uuid.UUID) ([]Order, error)
	Transition(context.Context, uuid.UUID, uuid.UUID, Event) (Order, error)
}
type PostgresRepo struct {
	pool          *pgxpool.Pool
	notifications NotificationWriter
}

func NewPostgresRepo(pool *pgxpool.Pool, notifications NotificationWriter) *PostgresRepo {
	return &PostgresRepo{pool: pool, notifications: notifications}
}

func (r *PostgresRepo) Create(ctx context.Context, o Order) (Order, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Order{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	now := time.Now().UTC()
	_, err = tx.Exec(ctx, `INSERT INTO store.orders(id,user_id,status,total_price,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$5)`, o.ID, o.UserID, o.Status, o.TotalPrice, now)
	if err != nil {
		return Order{}, err
	}
	for i := range o.Items {
		it := &o.Items[i]
		if it.ID == uuid.Nil {
			it.ID = uuid.New()
		}
		_, err = tx.Exec(ctx, `INSERT INTO store.order_items(id,order_id,product_id,store_id,quantity,price,product_name,store_name)
            VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, it.ID, o.ID, it.ProductID, it.StoreID, it.Quantity, it.Price, it.ProductName, it.StoreName)
		if err != nil {
			return Order{}, err
		}
	}
	_, err = tx.Exec(ctx, `INSERT INTO store.order_state_history(order_id,from_status,to_status,event) VALUES($1,NULL,$2,'created')`, o.ID, o.Status)
	if err != nil {
		return Order{}, err
	}
	if err = r.notifications.CreateTx(ctx, tx, o.UserID, o.ID, fmt.Sprintf("Заказ %s создан", o.ID)); err != nil {
		return Order{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Order{}, err
	}
	o.CreatedAt = now
	o.UpdatedAt = now
	o.History = []StateChange{{To: StatusCreated, Event: EventCreated, CreatedAt: now}}
	return o, nil
}
func (r *PostgresRepo) Get(ctx context.Context, id, userID uuid.UUID) (Order, error) {
	var o Order
	err := r.pool.QueryRow(ctx, `SELECT id,user_id,status,total_price,created_at,updated_at FROM store.orders WHERE id=$1 AND user_id=$2`, id, userID).
		Scan(&o.ID, &o.UserID, &o.Status, &o.TotalPrice, &o.CreatedAt, &o.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Order{}, ErrOrderNotFound
	}
	if err != nil {
		return Order{}, err
	}
	o.Items, err = r.items(ctx, id)
	if err != nil {
		return Order{}, err
	}
	o.History, err = r.history(ctx, id)
	return o, err
}
func (r *PostgresRepo) items(ctx context.Context, id uuid.UUID) ([]OrderItem, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,product_id,store_id,product_name,store_name,quantity,price
        FROM store.order_items WHERE order_id=$1 ORDER BY id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]OrderItem, 0)
	for rows.Next() {
		var it OrderItem
		if err := rows.Scan(&it.ID, &it.ProductID, &it.StoreID, &it.ProductName, &it.StoreName, &it.Quantity, &it.Price); err != nil {
			return nil, err
		}
		it.LineTotal = it.Price.Mul(decimal.NewFromInt(it.Quantity))
		items = append(items, it)
	}
	return items, rows.Err()
}
func (r *PostgresRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]Order, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,user_id,status,total_price,created_at,updated_at FROM store.orders WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Order, 0)
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.TotalPrice, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range out {
		out[i].Items, err = r.items(ctx, out[i].ID)
		if err != nil {
			return nil, err
		}
		out[i].History, err = r.history(ctx, out[i].ID)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}
func (r *PostgresRepo) history(ctx context.Context, id uuid.UUID) ([]StateChange, error) {
	rows, err := r.pool.Query(ctx, `SELECT from_status,to_status,event,created_at FROM store.order_state_history WHERE order_id=$1 ORDER BY created_at,id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	history := make([]StateChange, 0)
	for rows.Next() {
		var item StateChange
		var from *string
		var to, event string
		if err := rows.Scan(&from, &to, &event, &item.CreatedAt); err != nil {
			return nil, err
		}
		if from != nil {
			value := Status(*from)
			item.From = &value
		}
		item.To = Status(to)
		item.Event = Event(event)
		history = append(history, item)
	}
	return history, rows.Err()
}
func (r *PostgresRepo) Transition(ctx context.Context, id, userID uuid.UUID, event Event) (Order, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Order{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var current Status
	var owner uuid.UUID
	err = tx.QueryRow(ctx, `SELECT status,user_id FROM store.orders WHERE id=$1 FOR UPDATE`, id).Scan(&current, &owner)
	if errors.Is(err, pgx.ErrNoRows) {
		return Order{}, ErrOrderNotFound
	}
	if err != nil {
		return Order{}, err
	}
	if owner != userID {
		return Order{}, ErrForbidden
	}
	next, err := NextStatus(current, event)
	if err != nil {
		return Order{}, err
	}
	_, err = tx.Exec(ctx, `UPDATE store.orders SET status=$2,updated_at=now() WHERE id=$1`, id, next)
	if err != nil {
		return Order{}, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO store.order_state_history(order_id,from_status,to_status,event) VALUES($1,$2,$3,$4)`, id, current, next, event)
	if err != nil {
		return Order{}, err
	}
	if err = r.notifications.CreateTx(ctx, tx, owner, id, fmt.Sprintf("Статус заказа %s изменён: %s → %s", id, current, next)); err != nil {
		return Order{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Order{}, err
	}
	return r.Get(ctx, id, userID)
}
