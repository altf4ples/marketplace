package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/altf4ples/marketplace/internal/cart"
	"github.com/altf4ples/marketplace/internal/catalog"
	"github.com/altf4ples/marketplace/internal/notification"
	"github.com/altf4ples/marketplace/internal/order"
	"github.com/altf4ples/marketplace/internal/platform"
	"github.com/altf4ples/marketplace/internal/users"
	"github.com/google/uuid"
)

type productReader struct{ catalog *catalog.Service }

func (p productReader) GetProduct(ctx context.Context, id uuid.UUID) (order.Product, error) {
	product, err := p.catalog.GetProduct(ctx, id)
	if errors.Is(err, catalog.ErrProductNotFound) {
		return order.Product{}, order.ErrProductNotFound
	}
	if err != nil {
		return order.Product{}, err
	}
	return order.Product{ID: product.ID, StoreID: product.StoreID, Name: product.Name, StoreName: product.StoreName, Price: product.Price}, nil
}

type userReader struct{ svc *users.Service }

func (u userReader) Exists(ctx context.Context, id uuid.UUID) error {
	err := u.svc.Exists(ctx, id)
	if errors.Is(err, users.ErrNotFound) {
		return order.ErrUserNotFound
	}
	return err
}

type cartReader struct{ svc *cart.Service }

func (c cartReader) Take(ctx context.Context, id uuid.UUID) ([]order.CartItem, error) {
	items, err := c.svc.Take(ctx, id)
	if err != nil {
		return nil, err
	}
	out := make([]order.CartItem, 0, len(items))
	for _, it := range items {
		out = append(out, order.CartItem{ProductID: it.ProductID, Quantity: it.Quantity})
	}
	return out, nil
}
func (c cartReader) Restore(ctx context.Context, id uuid.UUID, items []order.CartItem) error {
	out := make([]cart.CartItem, 0, len(items))
	for _, it := range items {
		out = append(out, cart.CartItem{ProductID: it.ProductID, Quantity: it.Quantity})
	}
	return c.svc.Restore(ctx, id, out)
}
func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
func run() error {
	cfg, err := platform.Load()
	if err != nil {
		return err
	}
	logger := platform.NewLogger(cfg.LogLevel)
	startupCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pool, err := platform.OpenPostgres(startupCtx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	rdb, err := platform.OpenRedis(startupCtx, cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		return err
	}
	defer rdb.Close()

	userService := users.NewService(users.NewPostgresRepo(pool))
	catalogService := catalog.NewService(catalog.NewPostgresRepo(pool))
	cartService := cart.NewService(cart.NewCartRepo(rdb))
	notificationRepo := notification.NewPostgresRepo(pool)
	orderService := order.NewService(order.NewPostgresRepo(pool, notificationRepo), cartReader{cartService}, productReader{catalogService}, userReader{userService})
	notificationService := notification.NewService(notificationRepo)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			platform.WriteError(w, 503, "not_ready", "database unavailable")
			return
		}
		if err := rdb.Ping(ctx).Err(); err != nil {
			platform.WriteError(w, 503, "not_ready", "redis unavailable")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	users.RegisterRoutes(mux, users.NewHandler(userService))
	catalog.RegisterRoutes(mux, catalog.NewHandler(catalogService))
	cart.RegisterRoutes(mux, cart.NewHandler(cartService))
	order.RegisterRoutes(mux, order.NewHandler(orderService))
	notification.RegisterRoutes(mux, notification.NewHandler(notificationService))

	handler := platform.RequestLogger(logger)(platform.Recoverer(logger)(mux))
	srv := platform.NewHTTPServer(cfg.HTTPAddr, handler)
	serverErr := make(chan error, 1)
	go func() { logger.Info("http server starting", "addr", cfg.HTTPAddr); serverErr <- srv.ListenAndServe() }()
	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	select {
	case err := <-serverErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-sigCtx.Done():
	}
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		_ = srv.Close()
		return err
	}
	if err := <-serverErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	logger.Info("server stopped")
	return nil
}
