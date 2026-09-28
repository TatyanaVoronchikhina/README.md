package source

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"golang.org/x/text/encoding/charmap"
)

type Valute struct {
	Charcode string `xml:"CharCode"`
	Nominal  string `xml:"Nominal"`
	Value    string `xml:"Value"`
}
type ValCurs struct {
	XMLName xml.Name `xml:"ValCurs"`
	Date    string   `xml:"Date,attr"`
	Valute  []Valute `xml:"Valute"`
}

func GetRawCBR(ctx context.Context, filePath string, urlCBR string) (io.ReadCloser, error) {
	if filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("Error opening file: %v", err)
		}
		return file, nil
	}
	client := http.Client{
		Timeout: time.Second * 5,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlCBR, nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating request: %v", err)
	}
	req.Header.Set("accept", "application/xml")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error executing request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("Error executing request: %d %s", resp.StatusCode, resp.Status)
	}
	return resp.Body, nil
}

func ParseRawDataCBR(r io.Reader) (ValCurs, error) {
	decoder := xml.NewDecoder(charmap.Windows1251.NewDecoder().Reader(r))
	var xmlValCurs ValCurs
	if err := decoder.Decode(&xmlValCurs); err != nil {
		return xmlValCurs, fmt.Errorf("Error decoding XML: %v", err)
	}
	return xmlValCurs, nil
}
