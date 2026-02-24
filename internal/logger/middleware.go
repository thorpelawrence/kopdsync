package logger

import (
	"context"
	"log/slog"
	"net/http"
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqLogger := slog.With(
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
		)
		ctx := context.WithValue(r.Context(), LoggerContextKey, reqLogger)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
