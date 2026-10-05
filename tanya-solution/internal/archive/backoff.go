package archive

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"moneychange/internal/domain"
)

func (s *Store) backoffDir() string { return filepath.Join(s.dataDir, "backoff") }

// NextAllowed: zero — отсрочки нет.
func (s *Store) NextAllowed(source string) (time.Time, error) {
	if err := domain.ValidateSourceID(source); err != nil {
		return time.Time{}, err
	}
	b, err := os.ReadFile(filepath.Join(s.backoffDir(), source+".txt"))
	if errors.Is(err, os.ErrNotExist) {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("read backoff: %w", err)
	}
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(string(b)))
	if err != nil {
		return time.Time{}, fmt.Errorf("backoff file of %s: %w", source, err)
	}
	return t, nil
}

func (s *Store) Defer(source string, until time.Time) error {
	if err := domain.ValidateSourceID(source); err != nil {
		return err
	}
	return writeAtomic(s.backoffDir(), source+".txt", func(w io.Writer) error {
		_, err := fmt.Fprintln(w, until.UTC().Format(time.RFC3339))
		return err
	})
}
