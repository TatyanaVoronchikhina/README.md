package config

import (
	"errors"
	"fmt"
	"os"
)

type Config struct {
	DataDir string
}

func MakeDataDir(inputMode bool) (*Config, error) {
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}

	if inputMode && dataDir != "./data-demo" {
		return nil, errors.New("DATA_DIR must be set to ./data-demo in input mode\"")
	}

	err := os.MkdirAll(dataDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("MakeDataDir error: %v", err)
	}

	return &Config{DataDir: dataDir}, nil
}
