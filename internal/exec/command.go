// Package exec implements the three main commands for piphos: ping, pull, and push.
//
// Each command function handles flag parsing, provider initialization, and execution
// of the requested operation.
package exec

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/kappapee/piphos/internal/beacon"
	"github.com/kappapee/piphos/internal/tender"
	"github.com/kappapee/piphos/internal/validate"
)

// Ping detects the public IP address using the specified beacon provider.
// The beacon provider can be specified with the -beacon flag (default: "aws").
func Ping(ctx context.Context, args []string) (string, error) {
	fs := flag.NewFlagSet("ping", flag.ExitOnError)
	bs := fs.String("beacon", "aws", "which beacon provider to use")
	fs.Parse(args)
	if err := validate.Command(fs.NArg()); err != nil {
		return "", err
	}
	b, err := beacon.New(*bs)
	if err != nil {
		return "", fmt.Errorf("failed to create beacon %s: %w", *bs, err)
	}
	return b.Ping(ctx)
}

// Pull retrieves all hostname-to-IP mappings from the specified tender provider.
// The tender provider can be specified with the -tender flag (default: "gh").
// Requires PIPHOS_GITHUB_TOKEN environment variable for the "gh" provider.
func Pull(ctx context.Context, args []string) (map[string]string, error) {
	fs := flag.NewFlagSet("pull", flag.ExitOnError)
	ts := fs.String("tender", "gh", "which tender provider to use")
	fs.Parse(args)
	if err := validate.Command(fs.NArg()); err != nil {
		return nil, err
	}
	t, err := tender.New(*ts)
	if err != nil {
		return nil, fmt.Errorf("failed to create tender %s: %w", *ts, err)
	}
	return t.Pull(ctx)
}

// Push updates the current hostname's IP address in the specified tender provider.
// The beacon provider can be specified with the -beacon flag (default: "aws").
// The tender provider can be specified with the -tender flag (default: "gh").
// Requires PIPHOS_GITHUB_TOKEN environment variable for the "gh" provider.
// The hostname is automatically detected from the system.
func Push(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("push", flag.ExitOnError)
	ts := fs.String("tender", "gh", "which tender provider to use")
	bs := fs.String("beacon", "aws", "which tender provider to use")
	fs.Parse(args)
	if err := validate.Command(fs.NArg()); err != nil {
		return err
	}
	localHostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("failed to get system's hostname: %w", err)
	}
	b, err := beacon.New(*bs)
	if err != nil {
		return fmt.Errorf("failed to create beacon %s: %w", *bs, err)
	}
	t, err := tender.New(*ts)
	if err != nil {
		return fmt.Errorf("failed to create tender %s: %w", *ts, err)
	}
	publicIP, err := b.Ping(ctx)
	if err != nil {
		return err
	}
	return t.Push(ctx, localHostname, publicIP)
}

// Help displays the command-line usage information for piphos.
// It provides a comprehensive overview of available commands, their options,
// and practical usage examples to help users understand how to use the tool.
func Help() {
	fmt.Println("")
	fmt.Println("usage: piphos <command> [options]")
	fmt.Println("")
	fmt.Println("commands:")
	fmt.Println("  help                                      # print this help message")
	fmt.Println("  ping                                      # check public IP using a beacon")
	fmt.Println("  push                                      # push public IP to tender")
	fmt.Println("  pull                                      # pull stored hostname->IP map from tender")
	fmt.Println("")
	fmt.Println("examples:")
	fmt.Println("  piphos ping                               # use default beacon (aws)")
	fmt.Println("  piphos ping -beacon haz                   # use specific beacon")
	fmt.Println("  piphos push                               # push to default tender (gh)")
	fmt.Println("  piphos push -tender gh                    # push to specific tender")
	fmt.Println("  piphos push -tender gh -beacon haz        # push to specific tender using specific beacon")
	fmt.Println("  piphos pull                               # retrieve stored hostname->IP map from default tender (gh)")
	fmt.Println("  piphos pull -tender gh                    # retrieve stored hostname->IP map from specific tender")
	fmt.Println("")
	fmt.Println("available beacons:")
	fmt.Println("  aws                                       # https://checkip.amazonaws.com")
	fmt.Println("  haz (default)                             # https://ipv4.icanhazip.com")
	fmt.Println("")
	fmt.Println("available tenders:")
	fmt.Println("  gh (default)                              # GitHub Gists")
	fmt.Println("")
}
