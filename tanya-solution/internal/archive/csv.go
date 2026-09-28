package archive

import (
	"encoding/csv"
	"fmt"
	"moneychange/internal/config"
	"moneychange/internal/importer"
	"os"
	"path/filepath"
	"time"
)

//data/archive/YYYY-MM-DD/<source>-<UTC timestamp with nanoseconds>.csv

func SaveSnapshot(datadir config.Config, sourceName string, records []importer.RateRecord) error {
	date := records[0].Observed_at

	moscowZone := time.FixedZone("Europe/Moscow", 3*60*60)
	moscowTime := date.In(moscowZone)
	fileDate := moscowTime.Format("2006-01-02")

	utcTime := date.UTC()
	utcTimeStamp := utcTime.Format("20060102T150405.000000000Z")

	dirPath := filepath.Join(datadir.DataDir, "archive", fileDate)
	if err := os.MkdirAll(dirPath, 0777); err != nil {
		return fmt.Errorf("Failed to create datadir %s: %v", dirPath, err)
	}
	fileName := fmt.Sprintf("%s-%s.csv", sourceName, utcTimeStamp)
	fullPath := filepath.Join(dirPath, fileName)

	tmpFile, err := os.CreateTemp(dirPath, "temporary-*.tmp")
	if err != nil {
		return fmt.Errorf("Error creating temporary.tmp: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	writer := csv.NewWriter(tmpFile)
	header := []string{"observed_at", "source_at", "source", "kind", "bank", "city", "channel", "base", "quote",
		"buy_rate", "sell_rate", "reference_rate", "source_url"}
	if err = writer.Write(header); err != nil {
		return fmt.Errorf("Error writing header: %v", err)
	}

	for _, r := range records {
		row := []string{
			r.Observed_at.Format(time.RFC3339Nano),
			r.Source_at.Format(time.RFC3339Nano),
			r.Source,
			r.Kind,
			r.Bank,
			r.City,
			r.Channel,
			r.Base,
			r.Quote,
			r.Buy_rate,
			r.Sell_rate,
			r.Reference_rate,
			r.Source_url,
		}
		if err = writer.Write(row); err != nil {
			return fmt.Errorf("Error writing row: %v", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("Error writing data: %v", err)
	}
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("Error syncing file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("Error closing file: %v", err)
	}
	if err := os.Rename(tmpFile.Name(), fullPath); err != nil {
		return fmt.Errorf("Error renaming file: %v", err)
	}
	return nil
}
