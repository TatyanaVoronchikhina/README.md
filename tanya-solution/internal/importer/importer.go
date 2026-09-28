// Package importer — граница операции import: ошибки источников логируются здесь один раз.
package importer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"moneychange/internal/domain"
	"moneychange/internal/source"
)

// DefaultBackoff — отсрочка после 429 без Retry-After.
const DefaultBackoff = 15 * time.Minute

type SnapshotStore interface {
	SaveSnapshot(domain.Snapshot) (string, error)
}

type BackoffStore interface {
	NextAllowed(source string) (time.Time, error)
	Defer(source string, until time.Time) error
}

type Importer struct {
	Snapshots SnapshotStore
	Backoff   BackoffStore
	Now       func() time.Time
	Log       *slog.Logger
}

var ErrIncomplete = errors.New("import incomplete")

// Run: отказ источника не останавливает остальные; ошибка сохранения отсрочки прерывает импорт.
func (im Importer) Run(ctx context.Context, sources []source.Source) error {
	var failed int
	for _, src := range sources {
		ok, err := im.importOne(ctx, src)
		if err != nil {
			return err
		}
		if !ok {
			failed++
		}
	}
	if failed > 0 {
		return fmt.Errorf("%w: %d of %d sources", ErrIncomplete, failed, len(sources))
	}
	return nil
}

func (im Importer) importOne(ctx context.Context, src source.Source) (bool, error) {
	start := im.Now()
	log := im.Log.With(slog.String("source", src.ID()))
	since := func() int64 { return im.Now().Sub(start).Milliseconds() }

	next, err := im.Backoff.NextAllowed(src.ID())
	if err != nil {
		log.Error("backoff read failed", "error", err.Error(), "duration_ms", since())
		return false, nil
	}
	if start.Before(next) {
		log.Warn("source skipped: backoff", "next_allowed_at", next.UTC().Format(time.RFC3339))
		return false, nil
	}

	records, err := src.Fetch(ctx)
	if err != nil {
		attrs := []any{"error", err.Error(), "duration_ms", since()}
		var httpErr *source.HTTPError
		if errors.As(err, &httpErr) {
			attrs = append(attrs, "http_status", httpErr.StatusCode)
		}
		log.Error("source fetch failed", attrs...)
		if httpErr != nil && httpErr.StatusCode == 429 {
			return false, im.deferSource(log, src.ID(), httpErr.RetryAfter)
		}
		return false, nil
	}

	snap, err := domain.NewSnapshot(src.ID(), records)
	if err != nil {
		log.Error("source response rejected", "error", err.Error(), "duration_ms", since())
		return false, nil
	}
	path, err := im.Snapshots.SaveSnapshot(snap)
	if err != nil {
		log.Error("snapshot save failed", "error", err.Error(), "duration_ms", since())
		return false, nil
	}
	log.Info("source imported", "records", len(snap.Records), "file", path, "duration_ms", since())
	return true, nil
}

func (im Importer) deferSource(log *slog.Logger, id string, retryAfter time.Duration) error {
	if retryAfter <= 0 {
		retryAfter = DefaultBackoff
	}
	until := im.Now().Add(retryAfter)
	if err := im.Backoff.Defer(id, until); err != nil {
		log.Error("backoff save failed", "error", err.Error())
		return fmt.Errorf("save backoff of %s: %w", id, err)
	}
	log.Warn("source deferred", "next_allowed_at", until.UTC().Format(time.RFC3339))
	return nil
}
