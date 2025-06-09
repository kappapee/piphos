package main

import (
	"errors"
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
		return errors.New("empty token")
	}
	// Add provider-specific validation
	switch tender {
	case TenderGithub:
		if !strings.HasPrefix(token, "ghp_") &&
			!strings.HasPrefix(token, "gho_") &&
			!strings.HasPrefix(token, "github_pat_") {
			return errors.New("invalid GitHub token format")
		}
	case TenderGitlab:
		if !strings.HasPrefix(token, "glpat_") &&
			!strings.HasPrefix(token, "glpat-") {
			return errors.New("invalid GitLab token format")
		}
	}
	return nil
}
