package platform

import (
	"context"
	"log/slog"
)

// Используем приватные типы для ключей контекста, чтобы избежать коллизий.
type loggerKey struct{}
type requestIDKey struct{}

// WithLogger возвращает новый контекст с заданным логгером.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// LoggerFrom извлекает *slog.Logger из контекста.
// Если логгер не найден, возвращается логгер по умолчанию.
func LoggerFrom(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

// WithRequestID возвращает новый контекст с заданным request_id.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

// RequestIDFrom извлекает request_id из контекста.
func RequestIDFrom(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	return ""
}
