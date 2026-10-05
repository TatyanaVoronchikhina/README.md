package source

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"

	"moneychange/internal/domain"
)

const cbrXMLURL = "https://www.cbr.ru/scripts/XML_daily.asp"

type cbrValCurs struct {
	Date   string      `xml:"Date,attr"`
	Valute []cbrValute `xml:"Valute"`
}

type cbrValute struct {
	CharCode string `xml:"CharCode"`
	Nominal  string `xml:"Nominal"`
	Value    string `xml:"Value"`
}

// Windows-1251 объявлена в самом XML, поэтому декодер подключается через CharsetReader.
func parseCBRXML(r io.Reader, m Meta) ([]domain.RateRecord, error) {
	dec := xml.NewDecoder(r)
	dec.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		switch strings.ToLower(charset) {
		case "windows-1251", "cp1251":
			return charmap.Windows1251.NewDecoder().Reader(input), nil
		default:
			return nil, fmt.Errorf("unsupported charset %q", charset)
		}
	}
	var doc cbrValCurs
	if err := dec.Decode(&doc); err != nil {
		return nil, fmt.Errorf("decode XML: %w", err)
	}

	var sourceAt time.Time
	if doc.Date != "" {
		t, err := time.ParseInLocation("02.01.2006", doc.Date, domain.Moscow)
		if err != nil {
			return nil, fmt.Errorf("ValCurs Date %q: %w", doc.Date, err)
		}
		sourceAt = t
	}

	records := make([]domain.RateRecord, 0, len(doc.Valute))
	for _, v := range doc.Valute {
		rate, err := perUnit(v.Value, v.Nominal)
		if err != nil {
			return nil, fmt.Errorf("valute %s: %w", v.CharCode, err)
		}
		records = append(records, referenceRecord(m, domain.KindReference, strings.TrimSpace(v.CharCode), rate, sourceAt))
	}
	return records, nil
}
