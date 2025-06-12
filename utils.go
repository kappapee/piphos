package main

import (
	"fmt"
	"net"
	"strings"
)

func validateIP(ip string) error {
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("invalid IP address format: %s", ip)
	}
	return nil
}

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

func showUsage() {
	fmt.Println("usage: piphos <command> [options]")
	fmt.Println("")
	fmt.Println("commands:")
	fmt.Println("  check    check public IP using a beacon")
	fmt.Println("  push     push public IP to tender")
	fmt.Println("")
	fmt.Println("examples:")
	fmt.Println("  piphos check                    # use default beacon")
	fmt.Println("  piphos check -b aws             # use specific beacon")
	fmt.Println("  piphos push -t github           # push to specific tender")
	fmt.Println("  piphos push -t github -b haz    # push to specific tender using specific beacon")
}
