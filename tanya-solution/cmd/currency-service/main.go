package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"moneychange/internal/config"
	"moneychange/internal/logging"
	"moneychange/internal/source"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	cfg, err := config.Load()
	log := logging.New(cfg.LogLevel)
	if err != nil {
		log.Error("invalid configuration", "operation", "startup", "error", err.Error())
		return 1
	}
	if len(args) == 0 {
		usage(os.Stderr)
		log.Error("command is required", "operation", "startup")
		return 1
	}

	switch args[0] {
	case "calculate":
		return runCalculate(args[1:], log.With(slog.String("operation", "calculate")))
	case "import":
		return runImport(args[1:], cfg, log.With(slog.String("operation", "import")))
	case "convert":
		return runConvert(args[1:], cfg, log.With(slog.String("operation", "convert")))
	case "help", "-h", "--help":
		usage(os.Stdout)
		return 0
	default:
		usage(os.Stderr)
		log.Error("unknown command", "operation", "startup", "command", args[0])
		return 1
	}
}

func usage(w io.Writer) {
	fmt.Fprintf(w, `Использование: currency-service <команда> [флаги]

Команды:
  calculate --amount 100 --rate 80
      Сумма по заданному курсу: 8000.00
  import [--input ./cbr_daily.xml]
      Импорт источников из SOURCES (доступны: %s).
      --input разбирает сохранённый ответ ЦБ XML как demo-cbr; требует DATA_DIR=./data-demo.
  convert --from USD --to RUB --amount 100 --channel cash [--city Moscow] [--bank "Demo Bank B"]
      Лучшее свежее банковское предложение и сумма к получению.

Окружение: DATA_DIR (./data), SOURCES (cbr-xml), MAX_RATE_AGE (30m), LOG_LEVEL (INFO), EXTRA_CA_FILE.
`, strings.Join(source.IDs(), ", "))
}

// parseFlags: --help → (false, 0), ошибка → (false, 1).
func parseFlags(fs *flag.FlagSet, args []string, log *slog.Logger) (ok bool, code int) {
	fs.SetOutput(os.Stderr)
	err := fs.Parse(args)
	switch {
	case err == nil && fs.NArg() > 0:
		log.Error("unexpected arguments", "error", fmt.Sprint(fs.Args()))
		return false, 1
	case err == nil:
		return true, 0
	case errors.Is(err, flag.ErrHelp):
		return false, 0
	default:
		log.Error("invalid flags", "error", err.Error())
		return false, 1
	}
}
