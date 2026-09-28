package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"moneychange/internal/archive"
	"moneychange/internal/config"
	"moneychange/internal/importer"
	"moneychange/internal/source"
)

const demoCBRSource = "demo-cbr"

func runImport(args []string, cfg config.Config, log *slog.Logger) int {
	fs := flag.NewFlagSet("import", flag.ContinueOnError)
	input := fs.String("input", "", "сохранённый ответ ЦБ XML; только с DATA_DIR=./data-demo")
	if ok, code := parseFlags(fs, args, log); !ok {
		return code
	}

	sources, err := importSources(*input, cfg)
	if err != nil {
		log.Error("import setup failed", "error", err.Error())
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	store := archive.New(cfg.DataDir)
	im := importer.Importer{Snapshots: store, Backoff: store, Now: time.Now, Log: log}
	if err := im.Run(ctx, sources); err != nil {
		// Отказы уже залогированы importer'ом.
		return 1
	}
	return 0
}

func importSources(input string, cfg config.Config) ([]source.Source, error) {
	if input != "" {
		if !cfg.IsDemoDataDir() {
			return nil, errDemoDataDir
		}
		return []source.Source{source.NewCBRXMLFile(demoCBRSource, input, time.Now)}, nil
	}
	client, err := source.NewHTTPClient(cfg.ExtraCAFile)
	if err != nil {
		return nil, err
	}
	return source.Build(cfg.Sources, source.Deps{HTTP: client, Now: time.Now})
}

var errDemoDataDir = errors.New("import --input requires DATA_DIR=" + config.DemoDataDir)
