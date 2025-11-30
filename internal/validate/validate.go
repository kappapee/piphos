// Package validate provides input validation functions for piphos commands and data.
package validate

import (
	"fmt"
	"net"
)

// Command validates that no unexpected non-flag arguments were provided.
// Returns an error if nonFlagArgs is greater than zero.
func Command(nonFlagArgs int) error {
	if nonFlagArgs > 0 {
		return fmt.Errorf("%d unexpected argument(s)", nonFlagArgs)
	}
	return nil
}

// IP validates that the provided string is a valid IP address format.
// Returns an error if the IP address cannot be parsed.
func IP(ip string) error {
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("invalid IP address format for IP %s", ip)
	}
	return nil
}

// Token validates that the provided authentication token is non-empty.
// Returns an error if the token is empty.
func Token(token string) error {
	if token == "" {
		return fmt.Errorf("invalid token: token is empty")
	}
	return nil
}
