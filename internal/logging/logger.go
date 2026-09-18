package logging

import (
	"errors"
	"log/slog"
	"os"
	"strings"
)

func ParseSlogLevel(level string) (slog.Level, error) {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "DEBUG":
		return slog.LevelDebug, nil
	case "INFO":
		return slog.LevelInfo, nil
	case "WARN":
		return slog.LevelWarn, nil
	case "ERROR":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, errors.New(level + " is not a valid loglevel")
	}
}

func New(level slog.Level) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	return slog.New(handler).With(slog.String("service", "currency-service"))
}
