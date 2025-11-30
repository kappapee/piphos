package beacon

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/kappapee/piphos/internal/config"
	"github.com/kappapee/piphos/internal/validate"
)

// web implements the Beacon interface using HTTP-based IP detection services.
type web struct {
	baseURL string
	client  *http.Client
	headers map[string]string
	name    string
}

// newWeb creates a web beacon with the specified base URL.
func newWeb(baseURL, name string) *web {
	return &web{
		baseURL: baseURL,
		client:  &http.Client{Timeout: config.HTTPClientTimeout},
		headers: map[string]string{"User-Agent": config.PiphosUserAgent},
		name:    name,
	}
}

// Ping queries the web beacon service and returns the public IP address.
// The response is validated to ensure it contains a valid IP address.
func (b *web) Ping(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.baseURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request for beacon %s: %w", b.name, err)
	}
	for k, v := range b.headers {
		req.Header.Set(k, v)
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get response from beacon %s: %w", b.name, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to close response body: %v\n", err)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected response status from beacon %s: %d", b.name, resp.StatusCode)
	}
	limitedBody := io.LimitReader(resp.Body, config.MaxResponseBodySize)
	content, err := io.ReadAll(limitedBody)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	publicIP := strings.TrimSpace(string(content))
	if err = validate.IP(publicIP); err != nil {
		return "", err
	}
	return publicIP, nil
}
