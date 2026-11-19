// internal/platform/middleware.go
package platform

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const maxRequestIDLen = 64

func sanitizeRequestID(id string) string {
	if len(id) == 0 || len(id) > maxRequestIDLen {
		return ""
	}
	for _, ch := range id {
		// разрешаем только безопасный алфавит
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') || ch == '-' || ch == '_') {
			return ""
		}
	}
	return id
}

// RequestLogger обогащает контекст логгером с request_id и логирует каждый запрос.
func RequestLogger(baseLogger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// 1. Пробрасываем входящий X-Request-ID, если он есть.
			requestID := sanitizeRequestID(r.Header.Get("X-Request-ID"))
			if requestID == "" {
				requestID = uuid.NewString()
			}

			// 2. Отдаём его в ответе.
			w.Header().Set("X-Request-ID", requestID)

			// 3. Логгер с request_id кладём в контекст.
			reqLogger := baseLogger.With("request_id", requestID)
			ctx := WithLogger(r.Context(), reqLogger)
			ctx = WithRequestID(ctx, requestID)

			// 4. Оборачиваем ResponseWriter, чтобы получить статус-код.
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			start := time.Now()

			// 5. Логируем после выполнения хендлера.
			defer func() {
				reqLogger.Info("request",
					"method", r.Method,
					"path", r.URL.Path,
					"status", rw.status,
					"duration_ms", time.Since(start).Milliseconds(),
				)
			}()

			next.ServeHTTP(rw, r.WithContext(ctx))
		})
	}
}

// responseWriter перехватывает WriteHeader, чтобы знать итоговый статус.
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.status = code
	rw.wroteHeader = true
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.wroteHeader = true
	}
	return rw.ResponseWriter.Write(b)
}
