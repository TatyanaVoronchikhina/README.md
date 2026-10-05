// Package source — адаптеры источников: один запрос и нормализация, без логов и файлов.
package source

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"moneychange/internal/domain"
)

type Source interface {
	ID() string
	Fetch(ctx context.Context) ([]domain.RateRecord, error)
}

type Meta struct {
	SourceID   string
	SourceURL  string
	ObservedAt time.Time
}

type ParseFunc func(r io.Reader, m Meta) ([]domain.RateRecord, error)

type Deps struct {
	HTTP *http.Client
	Now  func() time.Time
}

// Новый источник — строка здесь и файл с ParseFunc.
var registry = map[string]struct {
	url, accept string
	parse       ParseFunc
}{
	"cbr-xml":           {cbrXMLURL, "application/xml", parseCBRXML},
	"cbr-json-mirror":   {cbrJSONURL, "application/json", parseCBRJSON},
	"moex-iss":          {moexURL, "application/json", parseMOEX},
	"tbank-public":      {tbankURL, "application/json", parseTBank},
	"raiffeisen-public": {raiffeisenURL, "application/json", parseRaiffeisen},
}

func IDs() []string {
	ids := make([]string, 0, len(registry))
	for id := range registry {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func Build(ids []string, d Deps) ([]Source, error) {
	seen := make(map[string]bool, len(ids))
	sources := make([]Source, 0, len(ids))
	for _, id := range ids {
		spec, ok := registry[id]
		if !ok {
			return nil, fmt.Errorf("unknown source %q, known: %v", id, IDs())
		}
		if seen[id] {
			return nil, fmt.Errorf("source %q is listed twice", id)
		}
		seen[id] = true
		sources = append(sources, &adapter{id: id, url: spec.url, open: HTTPGet(d.HTTP, spec.url, spec.accept), parse: spec.parse, now: d.Now})
	}
	return sources, nil
}

// NewCBRXMLFile — ЦБ XML из файла для import --input.
func NewCBRXMLFile(id, path string, now func() time.Time) Source {
	return &adapter{id: id, url: path, open: File(path), parse: parseCBRXML, now: now}
}

type adapter struct {
	id    string
	url   string
	open  Opener
	parse ParseFunc
	now   func() time.Time
}

func (a *adapter) ID() string { return a.id }

func (a *adapter) Fetch(ctx context.Context) ([]domain.RateRecord, error) {
	body, err := a.open(ctx)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	meta := Meta{SourceID: a.id, SourceURL: a.url, ObservedAt: a.now().UTC()}
	records, err := a.parse(body, meta)
	if err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return records, nil
}
