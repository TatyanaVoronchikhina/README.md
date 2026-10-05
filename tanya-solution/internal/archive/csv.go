// Package archive хранит снимки источников в CSV: data/archive/YYYY-MM-DD/<source>-<UTC ts>.csv.
package archive

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"moneychange/internal/domain"
)

var header = []string{"observed_at", "source_at", "source", "kind", "bank", "city", "channel",
	"base", "quote", "buy_rate", "sell_rate", "reference_rate", "source_url"}

const fileTimeLayout = "20060102T150405.000000000Z"

type Store struct {
	dataDir string
}

func New(dataDir string) *Store {
	return &Store{dataDir: dataDir}
}

func (s *Store) archiveDir() string { return filepath.Join(s.dataDir, "archive") }

// SaveSnapshot пишет атомарно и сверяет с обратным чтением. Каталог дня — observed_at в Europe/Moscow.
func (s *Store) SaveSnapshot(snap domain.Snapshot) (string, error) {
	dir := filepath.Join(s.archiveDir(), domain.MoscowDate(snap.ObservedAt))
	name := fmt.Sprintf("%s-%s.csv", snap.Source, snap.ObservedAt.UTC().Format(fileTimeLayout))
	path := filepath.Join(dir, name)

	err := writeAtomic(dir, name, func(w io.Writer) error {
		cw := csv.NewWriter(w)
		if err := cw.Write(header); err != nil {
			return err
		}
		for _, r := range snap.Records {
			if err := cw.Write(encode(r)); err != nil {
				return err
			}
		}
		cw.Flush()
		return cw.Error()
	})
	if err != nil {
		return "", fmt.Errorf("save snapshot %s: %w", name, err)
	}

	back, err := ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read back snapshot %s: %w", name, err)
	}
	if len(back) != len(snap.Records) {
		return "", fmt.Errorf("read back snapshot %s: %d of %d records", name, len(back), len(snap.Records))
	}
	for i := range back {
		if !slices.Equal(encode(back[i]), encode(snap.Records[i])) {
			return "", fmt.Errorf("read back snapshot %s: record %d differs", name, i+1)
		}
	}
	return path, nil
}

// ReadSince: временные файлы пропускаются, повреждённый CSV — ошибка.
func (s *Store) ReadSince(since time.Time) ([]domain.RateRecord, error) {
	days, err := os.ReadDir(s.archiveDir())
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read archive: %w", err)
	}
	from := domain.MoscowDate(since)

	var records []domain.RateRecord
	for _, day := range days {
		if !day.IsDir() || !isDate(day.Name()) || day.Name() < from {
			continue
		}
		dir := filepath.Join(s.archiveDir(), day.Name())
		files, err := os.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("read archive day %s: %w", day.Name(), err)
		}
		for _, f := range files {
			if f.IsDir() || !isSnapshotFile(f.Name()) {
				continue
			}
			recs, err := ReadFile(filepath.Join(dir, f.Name()))
			if err != nil {
				return nil, err
			}
			records = append(records, recs...)
		}
	}
	return records, nil
}

func ReadFile(path string) ([]domain.RateRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open snapshot: %w", err)
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("snapshot %s: %w", filepath.Base(path), err)
	}
	if len(rows) == 0 || !slices.Equal(rows[0], header) {
		return nil, fmt.Errorf("snapshot %s: unexpected header", filepath.Base(path))
	}
	records := make([]domain.RateRecord, 0, len(rows)-1)
	for i, row := range rows[1:] {
		r, err := decode(row)
		if err == nil {
			err = r.Validate()
		}
		if err != nil {
			return nil, fmt.Errorf("snapshot %s line %d: %w", filepath.Base(path), i+2, err)
		}
		records = append(records, r)
	}
	return records, nil
}

func encode(r domain.RateRecord) []string {
	return []string{
		r.ObservedAt.UTC().Format(time.RFC3339Nano),
		formatTime(r.SourceAt),
		r.Source,
		string(r.Kind),
		r.Bank,
		r.City,
		string(r.Channel),
		r.Base,
		r.Quote,
		formatDecimal(r.BuyRate),
		formatDecimal(r.SellRate),
		formatDecimal(r.ReferenceRate),
		r.SourceURL,
	}
}

func decode(row []string) (domain.RateRecord, error) {
	var (
		r   domain.RateRecord
		err error
	)
	if r.ObservedAt, err = time.Parse(time.RFC3339Nano, row[0]); err != nil {
		return r, fmt.Errorf("observed_at: %w", err)
	}
	if r.SourceAt, err = parseTime(row[1]); err != nil {
		return r, fmt.Errorf("source_at: %w", err)
	}
	r.Source, r.Kind, r.Bank, r.City, r.Channel = row[2], domain.Kind(row[3]), row[4], row[5], domain.Channel(row[6])
	r.Base, r.Quote, r.SourceURL = row[7], row[8], row[12]
	for i, dst := range []*decimal.NullDecimal{&r.BuyRate, &r.SellRate, &r.ReferenceRate} {
		if *dst, err = parseDecimal(row[9+i]); err != nil {
			return r, fmt.Errorf("%s: %w", header[9+i], err)
		}
	}
	return r, nil
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339Nano)
}

func parseTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339Nano, s)
}

func formatDecimal(d decimal.NullDecimal) string {
	if !d.Valid {
		return ""
	}
	return d.Decimal.String()
}

func parseDecimal(s string) (decimal.NullDecimal, error) {
	if s == "" {
		return decimal.NullDecimal{}, nil
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.NullDecimal{}, err
	}
	return decimal.NewNullDecimal(d), nil
}

func isDate(name string) bool {
	_, err := time.Parse(time.DateOnly, name)
	return err == nil
}

func isSnapshotFile(name string) bool {
	return strings.HasSuffix(name, ".csv") && !strings.HasPrefix(name, ".")
}

// writeAtomic: temp в той же директории → Sync → Close → rename; при ошибке temp удаляется.
func writeAtomic(dir, name string, write func(io.Writer) error) (err error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}
	tmp, err := os.CreateTemp(dir, "."+name+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer func() {
		if err != nil {
			tmp.Close()
			os.Remove(tmp.Name())
		}
	}()

	if err = write(tmp); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	if err = tmp.Chmod(0o644); err != nil {
		return fmt.Errorf("chmod: %w", err)
	}
	if err = tmp.Sync(); err != nil {
		return fmt.Errorf("sync: %w", err)
	}
	if err = tmp.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}
	if err = os.Rename(tmp.Name(), filepath.Join(dir, name)); err != nil {
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}
