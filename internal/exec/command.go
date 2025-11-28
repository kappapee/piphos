package exec

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/kappapee/piphos/internal/beacon"
	"github.com/kappapee/piphos/internal/tender"
)

func Ping(ctx context.Context, args []string) (string, error) {
	fs := flag.NewFlagSet("ping", flag.ExitOnError)
	bs := fs.String("beacon", "haz", "which beacon provider to use")
	fs.Parse(args)
	b, err := beacon.New(*bs)
	if err != nil {
		return "", fmt.Errorf("failed to create beacon %s: %w", *bs, err)
	}
	return b.Ping(ctx)
}

func Pull(ctx context.Context, args []string) (map[string]string, error) {
	fs := flag.NewFlagSet("pull", flag.ExitOnError)
	ts := fs.String("tender", "gh", "which tender provider to use")
	fs.Parse(args)
	t, err := tender.New(*ts)
	if err != nil {
		return nil, fmt.Errorf("failed to create tender %s: %w", *ts, err)
	}
	return t.Pull(ctx)
}

func Push(ctx context.Context, args []string, ip string) error {
	fs := flag.NewFlagSet("push", flag.ExitOnError)
	ts := fs.String("tender", "gh", "which tender provider to use")
	fs.Parse(args)
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("failed to get system's hostname: %w", err)
	}
	t, err := tender.New(*ts)
	if err != nil {
		return fmt.Errorf("failed to create tender %s: %w", *ts, err)
	}
	return t.Push(ctx, hostname, ip)
}
