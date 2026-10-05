package logging

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

func ParseLevel(str string) (slog.Level, error) {
	switch strings.TrimSpace(strings.ToUpper(str)) {
	case "DEBUG":
		return slog.LevelDebug, nil
	case "INFO":
		return slog.LevelInfo, nil
	case "WARN":
		return slog.LevelWarn, nil
	case "ERROR":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("%q is not a valid log level", str)
	}
}

func New(level slog.Level) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	return slog.New(handler).With(slog.String("service", "currency-service"))
}
