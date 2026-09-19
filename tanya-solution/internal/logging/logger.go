package logging

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

func SlogLvl(str string) (slog.Level, error) {
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
		return slog.LevelInfo, fmt.Errorf(str + " is not a valid loglevel")
	}
}

func New(slogLevel slog.Level) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{ //почему в терминал, удобнее же в отдельный файл?
		Level: slogLevel,
	})
	return slog.New(handler)
}
