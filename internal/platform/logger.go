package platform

import (
	"log/slog"
	"os"
)

// NewLogger создает и настраивает новый *slog.Logger.
// Для продакшена обычно используют JSON, для локальной разработки — Text.
func NewLogger(level slog.Level) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: level,
	}
	// В реальном проекте можно выбирать хендлер в зависимости от окружения.
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	// Устанавливаем логгер по умолчанию. Теперь slog.Info/Error и т.д.
	// будут использовать этот логгер.
	slog.SetDefault(logger)

	return logger
}
