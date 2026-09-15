package main

import (
	"errors"
	"flag"
	"fmt"
	"moneychange/internal/domain"
	"os"
)

func main() {
	flag.Usage = func() {
		fmt.Println("Верный формат ввода для конвертации: `calculate --amount 100 --rate 80`")
	}
	if len(os.Args) < 2 {
		flag.Usage()
		errByLenArgs := errors.New("not enough data")
		fmt.Fprint(os.Stderr, "CLI-command error", errByLenArgs)
		os.Exit(1)
	}
	switch os.Args[1] {
	case "calculate":
		calcFlag := flag.NewFlagSet("calculate", flag.ContinueOnError)
		amount := calcFlag.String("amount", "", "amount to calculate")
		rate := calcFlag.String("rate", "", "rate to calculate")
		if err := calcFlag.Parse(os.Args[2:]); err != nil {
			fmt.Fprint(os.Stderr, "parsing error `calculate`", err)
			os.Exit(1)
		}
		result, err := domain.Result(*amount, *rate)
		if err != nil {
			fmt.Fprint(os.Stderr, "convertation error", err)
			os.Exit(1)
		}
		fmt.Println(result)
	case "--help", "-h", "help":
		flag.Usage()
	default:
		otherErr := errors.New("unknown command " + os.Args[1])
		fmt.Fprint(os.Stderr, "unknown error ", otherErr)
		os.Exit(1)
	}

}
