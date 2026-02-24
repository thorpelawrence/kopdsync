package logger

import (
	"context"
	"log/slog"
)

type contextKey string

const LoggerContextKey = contextKey("logger")

func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(LoggerContextKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}
