package main

import (
	"fmt"
	"net"
	"strings"
)

// validateIP verifies that the provided string represents a valid IP address.
//
// Parameters:
//   - ip: The IP address string to validate
//
// Returns:
//   - error: An error if the IP address format is invalid, nil if valid
func validateIP(ip string) error {
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("invalid IP address format: %s", ip)
	}
	return nil
}

// validateToken verifies that the provided authentication token is valid
// for the specified tender service. Different tender services have different
// token format requirements and validation rules.
//
// Parameters:
//   - token: The authentication token to validate
//   - tender: The tender service identifier that will use this token
//
// Returns:
//   - error: An error if the token is empty, malformed, or incompatible
//     with the specified tender service, nil if valid
func validateToken(token, tender string) error {
	if token == "" {
		return fmt.Errorf("empty token")
	}
	switch tender {
	case TenderGithub:
		if !strings.HasPrefix(token, "ghp_") &&
			!strings.HasPrefix(token, "gho_") &&
			!strings.HasPrefix(token, "github_pat_") {
			return fmt.Errorf("invalid GitHub token format")
		}
	}
	return nil
}

func validateCmd(countNonFlagArgs int) error {
	if countNonFlagArgs > 0 {
		return fmt.Errorf("found %d unexpected arguments", countNonFlagArgs)
	}
	return nil
}

// showUsage displays the command-line usage information for piphos.
// It provides a comprehensive overview of available commands, their options,
// and practical usage examples to help users understand how to use the tool.
func showUsage() {
	fmt.Println("usage: piphos <command> [options]")
	fmt.Println("")
	fmt.Println("commands:")
	fmt.Println("  check    check public IP using a beacon")
	fmt.Println("  push     push public IP to tender")
	fmt.Println("  pull     pull stored IPs from tender")
	fmt.Println("")
	fmt.Println("examples:")
	fmt.Println("  piphos check                    # use default beacon")
	fmt.Println("  piphos check -b aws             # use specific beacon")
	fmt.Println("  piphos push -t github           # push to specific tender")
	fmt.Println("  piphos push -t github -b haz    # push to specific tender using specific beacon")
	fmt.Println("  piphos pull -t github           # retrieve stored IPs from tender")
}
