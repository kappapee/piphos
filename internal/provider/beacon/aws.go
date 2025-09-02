package beacon

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/kappapee/piphos/internal/service"
)

type BeaconAWS service.Beacon

func NewBeaconAWS() *BeaconAWS {
	return &BeaconAWS{
		Name: "aws",
		URL:  "https://checkip.amazonaws.com",
	}
}

func (b *BeaconAWS) Ping(ctx context.Context, client *http.Client) (string, error) {
	url := b.URL
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("unable to create request: %w", err)
	}
	req.Header.Set("User-Agent", "piphos/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("unable to get response: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "WARN: unable to close response body: %v\n", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("beacon returned status %d: %s", resp.StatusCode, string(body))
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("unable to read response body: %w", err)
	}

	return strings.TrimSpace(string(content)), nil
}
