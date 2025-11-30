// Package beacon provides interfaces and implementations for detecting public IP addresses.
//
// The Beacon interface defines a strategy for IP detection, allowing different
// beacon services to be used interchangeably. Current implementations include
// icanhazip.com ("haz") and Amazon's checkip service ("aws").
package beacon

import (
	"context"
	"fmt"
)

// Beacon defines the interface for IP address detection services.
// Implementations provide different strategies for discovering the public IP address.
type Beacon interface {
	// Ping detects and returns the public IP address.
	// Returns an error if the beacon service is unreachable or returns invalid data.
	Ping(ctx context.Context) (string, error)
}

// New creates a Beacon instance for the specified provider.
// Supported providers are "haz" (icanhazip.com) and "aws" (Amazon checkip).
// Returns an error if the provider is unknown.
func New(beacon string) (Beacon, error) {
	switch beacon {
	case "haz":
		return newWeb("https://ipv4.icanhazip.com", "haz"), nil
	case "aws":
		return newWeb("https://checkip.amazonaws.com", "aws"), nil
	default:
		return nil, fmt.Errorf("unknown beacon: %s", beacon)
	}
}
