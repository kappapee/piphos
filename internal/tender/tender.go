// Package tender provides interfaces and implementations for storing hostname-to-IP mappings.
//
// The Tender interface defines a storage strategy with Pull (retrieve) and Push (update)
// operations. The primary implementation uses GitHub Gists ("gh") as a backend, storing
// mappings in a private gist identified by the description "_piphos_".
package tender

import (
	"context"
	"fmt"
	"os"

	"github.com/kappapee/piphos/internal/validate"
)

// Tender defines the interface for storing and retrieving hostname-to-IP mappings.
type Tender interface {
	// Pull retrieves all hostname-to-IP mappings from storage.
	// Returns a map where keys are hostnames and values are IP addresses.
	Pull(ctx context.Context) (map[string]string, error)

	// Push stores or updates the IP address for a given hostname.
	// If the hostname already exists with the same IP, no update is performed.
	Push(ctx context.Context, hostname, ip string) error
}

// New creates a Tender instance for the specified provider.
// Supported providers are "gh" (GitHub Gists, requires GITHUB_TOKEN environment variable).
// Returns an error if the provider is unknown or required credentials are missing.
func New(tender string) (Tender, error) {
	switch tender {
	case "gh":
		token := os.Getenv("GITHUB_TOKEN")
		err := validate.Token(token)
		if err != nil {
			return nil, err
		}
		return newGithub(token), nil
	default:
		return nil, fmt.Errorf("unknown tender: %s", tender)
	}
}
