// Piphos tracks dynamic IP addresses using GitHub Gists as storage.
//
// Usage:
//
//	piphos ping [-beacon=PROVIDER]    # Detect public IP
//	piphos pull [-tender=PROVIDER]    # Retrieve all tracked hosts
//	piphos push [-tender=PROVIDER]    # Update current hostname's IP
//
// The push and pull commands require the PIPHOS_GITHUB_TOKEN environment variable.
//
// Examples:
//
//	export PIPHOS_GITHUB_TOKEN=ghp_xxx
//	piphos ping                    # 203.0.113.42
//	piphos push                    # 203.0.113.42
//	piphos pull                    # laptop: 203.0.113.42
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kappapee/piphos/internal/exec"
)

func main() {
	if len(os.Args) < 2 {
		exec.Help()
		os.Exit(1)
	}
	ctx := context.Background()
	switch os.Args[1] {
	case "ping":
		publicIP, err := exec.Ping(ctx, os.Args[2:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to run ping command: %v\n", err)
			exec.Help()
			os.Exit(1)
		}
		fmt.Fprintln(os.Stdout, publicIP)
	case "pull":
		hosts, err := exec.Pull(ctx, os.Args[2:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to run pull command: %v\n", err)
			exec.Help()
			os.Exit(1)
		}
		for k, v := range hosts {
			fmt.Fprintf(os.Stdout, "%s: %s\n", k, v)
		}
	case "push":
		publicIP, err := exec.Ping(ctx, []string{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to check public IP address: %v\n", err)
			os.Exit(1)
		}
		if err := exec.Push(ctx, os.Args[2:], publicIP); err != nil {
			fmt.Fprintf(os.Stderr, "failed to run push command: %v\n", err)
			exec.Help()
			os.Exit(1)
		}
		fmt.Fprintln(os.Stdout, publicIP)
	case "help":
		exec.Help()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %v\n", os.Args[1])
		exec.Help()
		os.Exit(1)
	}
}
