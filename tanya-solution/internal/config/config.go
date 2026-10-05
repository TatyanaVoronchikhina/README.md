package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"moneychange/internal/logging"
)

const DemoDataDir = "./data-demo"

type Config struct {
	DataDir    string
	Sources    []string
	MaxRateAge time.Duration
	LogLevel   slog.Level
	// ExtraCAFile — PEM с доп. корнями, например Russian Trusted Root CA для www.tbank.ru.
	ExtraCAFile string
}

// Load возвращает все ошибки сразу; LogLevel заполнен и при ошибке, чтобы main мог её залогировать.
func Load() (Config, error) {
	cfg := Config{
		DataDir:     getenv("DATA_DIR", "./data"),
		Sources:     splitList(getenv("SOURCES", "cbr-xml")),
		LogLevel:    slog.LevelInfo,
		ExtraCAFile: os.Getenv("EXTRA_CA_FILE"),
	}
	var errs []error

	level, err := logging.ParseLevel(getenv("LOG_LEVEL", "INFO"))
	if err != nil {
		errs = append(errs, fmt.Errorf("LOG_LEVEL: %w", err))
	} else {
		cfg.LogLevel = level
	}

	cfg.MaxRateAge, err = time.ParseDuration(getenv("MAX_RATE_AGE", "30m"))
	if err != nil || cfg.MaxRateAge <= 0 {
		errs = append(errs, fmt.Errorf("MAX_RATE_AGE must be a positive duration like 30m"))
	}
	if len(cfg.Sources) == 0 {
		errs = append(errs, errors.New("SOURCES must list at least one source"))
	}
	return cfg, errors.Join(errs...)
}

func (c Config) IsDemoDataDir() bool {
	return filepath.Base(filepath.Clean(c.DataDir)) == filepath.Base(DemoDataDir)
}

func getenv(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func splitList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}
