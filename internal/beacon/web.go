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

type web struct {
	baseURL string
	client  *http.Client
	headers map[string]string
	name    string
}

func newWeb(baseURL, name string) *web {
	return &web{
		baseURL: baseURL,
		client:  &http.Client{Timeout: config.HTTPClientTimeout},
		headers: map[string]string{"User-Agent": config.PiphosUserAgent},
		name:    name,
	}
}

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
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response from beacon %s: %w", b.name, err)
	}
	publicIP := strings.TrimSpace(string(content))
	if err = validate.IP(publicIP); err != nil {
		return "", err
	}
	return publicIP, nil
}
