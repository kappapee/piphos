package beacon

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/kappapee/piphos/internal/config"
	"github.com/kappapee/piphos/internal/validation"
)

type Beacon interface {
	Ping(ctx context.Context) (string, error)
}

type HTTPBeacon struct {
	baseURL string
	client  *http.Client
	name    string
}

func NewHTTPBeacon(baseURL, name string) *HTTPBeacon {
	return &HTTPBeacon{
		baseURL: baseURL,
		client:  &http.Client{Timeout: config.HTTPClientTimeout},
		name:    name,
	}
}

func New(beacon string) (*HTTPBeacon, error) {
	switch beacon {
	case "haz":
		return NewHTTPBeacon("https://ipv4.icanhazip.com", "haz"), nil
	case "aws":
		return NewHTTPBeacon("https://checkip.amazonaws.com", "aws"), nil
	default:
		return nil, fmt.Errorf("unknown beacon: %s", beacon)
	}
}

func (b *HTTPBeacon) Ping(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.baseURL, nil)
	if err != nil {
		return "", fmt.Errorf("unable to create request for beacon %s: %w", b.name, err)
	}
	req.Header.Set("User-Agent", "piphos/1.0")
	resp, err := b.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("unable to get response from beacon %s: %w", b.name, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "unable to close response body: %v\n", err)
		}
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("beacon %s returned status %d: %s", b.name, resp.StatusCode, string(body))
	}
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("unable to read response body from beacon %s: %w", b.name, err)
	}
	publicIP := strings.TrimSpace(string(content))
	if err = validation.IP(publicIP); err != nil {
		return "", fmt.Errorf("invalid IP address format: %w", err)
	}
	return publicIP, nil
}
