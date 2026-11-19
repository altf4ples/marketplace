package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/altf4ples/marketplace/internal/cart"
	"github.com/altf4ples/marketplace/internal/users"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func TestUserRepositoryPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if _, err := pool.Exec(ctx, `CREATE SCHEMA IF NOT EXISTS store`); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS store.users (id UUID PRIMARY KEY, username VARCHAR(50) NOT NULL UNIQUE, phone_number VARCHAR(15) UNIQUE)`); err != nil {
		t.Fatal(err)
	}

	repo := users.NewPostgresRepo(pool)
	user := users.User{ID: uuid.New(), Name: "integration-" + uuid.NewString()}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM store.users WHERE id=$1`, user.ID) })
	if err := repo.Exists(ctx, user.ID); err != nil {
		t.Fatalf("Exists: %v", err)
	}
	got, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != user.ID || got.Name != user.Name {
		t.Fatalf("got %#v, want %#v", got, user)
	}
}

func TestCartRepositoryRedis(t *testing.T) {
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("TEST_REDIS_ADDR is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client := redis.NewClient(&redis.Options{Addr: addr, Password: os.Getenv("TEST_REDIS_PASSWORD")})
	defer client.Close()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatal(err)
	}

	repo := cart.NewCartRepo(client)
	userID, productID := uuid.New(), uuid.New()
	t.Cleanup(func() { _ = repo.Clear(context.Background(), userID) })
	if _, err := repo.AddItem(ctx, userID, productID, 2); err != nil {
		t.Fatal(err)
	}
	items, err := repo.Take(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ProductID != productID || items[0].Quantity != 2 {
		t.Fatalf("unexpected cart snapshot: %#v", items)
	}
	remaining, err := repo.Get(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining.Items) != 0 {
		t.Fatalf("cart was not cleared by Take: %#v", remaining.Items)
	}
}
