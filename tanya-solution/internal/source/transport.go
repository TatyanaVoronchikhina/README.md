package source

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	RequestTimeout = 5 * time.Second
	userAgent      = "currency-service/0.4"
	maxBodyBytes   = 10 << 20
)

type Opener func(ctx context.Context) (io.ReadCloser, error)

// HTTPError — статус не 200; RetryAfter — из заголовка Retry-After.
type HTTPError struct {
	StatusCode int
	RetryAfter time.Duration
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("unexpected HTTP status %d %s", e.StatusCode, http.StatusText(e.StatusCode))
}

// NewHTTPClient: extraCAFile добавляет корневые сертификаты к системным.
func NewHTTPClient(extraCAFile string) (*http.Client, error) {
	client := &http.Client{Timeout: RequestTimeout}
	if extraCAFile == "" {
		return client, nil
	}
	pem, err := os.ReadFile(extraCAFile)
	if err != nil {
		return nil, fmt.Errorf("read EXTRA_CA_FILE: %w", err)
	}
	pool, err := x509.SystemCertPool()
	if err != nil || pool == nil {
		pool = x509.NewCertPool()
	}
	if !pool.AppendCertsFromPEM(pem) {
		return nil, errors.New("EXTRA_CA_FILE contains no PEM certificates")
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12}
	client.Transport = transport
	return client, nil
}

func HTTPGet(client *http.Client, url, accept string) Opener {
	return func(ctx context.Context) (io.ReadCloser, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Accept", accept)
		req.Header.Set("User-Agent", userAgent)

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request %s: %w", url, err)
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, &HTTPError{StatusCode: resp.StatusCode, RetryAfter: parseRetryAfter(resp.Header.Get("Retry-After"), time.Now())}
		}
		return limitedBody{Reader: io.LimitReader(resp.Body, maxBodyBytes), Closer: resp.Body}, nil
	}
}

func File(path string) Opener {
	return func(context.Context) (io.ReadCloser, error) {
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("open input: %w", err)
		}
		return f, nil
	}
}

type limitedBody struct {
	io.Reader
	io.Closer
}

// parseRetryAfter: секунды или HTTP-дата.
func parseRetryAfter(v string, now time.Time) time.Duration {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil && secs > 0 {
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil && t.After(now) {
		return t.Sub(now)
	}
	return 0
}
