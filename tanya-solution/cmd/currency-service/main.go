package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"moneychange/internal/domain"
	"moneychange/internal/logging"
	"os"
	"time"
)

// git checkout main
// git pull
// git branch fitcher/t.voronchikhina/task-2-add-logger
// git checkout fitcher/t.voronchikhina/task-2-add-logger

func main() {
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "INFO"
	}
	slogLvl, err := logging.SlogLvl(level)
	if err != nil {
		accidentLogger := logging.New(slogLvl)
		accidentLogger.Error("invalid logLevel ", "level", level, "error", err)
		os.Exit(1)
	}
	baseLogger := logging.New(slogLvl)
	logger := baseLogger.With(slog.String("service", "currency-service"))

	flag.Usage = func() {
		fmt.Fprintln(os.Stdout, "Верный формат ввода для конвертации: `calculate --amount 100 --rate 80`")
	}
	if len(os.Args) < 2 {
		flag.Usage()
		errByLenArgs := errors.New("not enough data")
		fmt.Fprintln(os.Stderr, "CLI-command error", errByLenArgs)
		os.Exit(1)
	}
	switch os.Args[1] {
	case "calculate":
		opLogger := logger.With(slog.String("operation", "calculate"))
		startTime := time.Now()
		calcFlag := flag.NewFlagSet("calculate", flag.ContinueOnError)
		amount := calcFlag.String("amount", "", "amount to calculate")
		rate := calcFlag.String("rate", "", "rate to calculate")
		if err := calcFlag.Parse(os.Args[2:]); err != nil {
			duration := time.Since(startTime).Milliseconds()
			opLogger.Error("calculate parsing error", "error", err.Error(), "duration_ms", duration, "source", "CLI-command")
			os.Exit(1)
		}

		result, err := domain.Convert(*amount, *rate, opLogger)
		duration := time.Since(startTime).Milliseconds()
		if err != nil {
			opLogger.Error("convertation error", "error", err.Error(), "duration_ms", duration, "source", "CLI-command")
			os.Exit(1)
		}
		opLogger.Info("calculation successfull", "records", 1, "duration_ms", duration, "source", "CLI-command")

		fmt.Println(result)

	case "--help", "-h", "help":
		flag.Usage()

	default:
		otherErr := errors.New("unknown command " + os.Args[1])
		logger.With(slog.String("operation", "invalid")).Error("unknown command", otherErr.Error())
		os.Exit(1)
	}

}
