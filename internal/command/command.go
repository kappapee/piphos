package command

import (
	"context"
	"flag"
	"fmt"

	"github.com/kappapee/piphos/internal/beacon"
)

func Ping(ctx context.Context, args []string) (string, error) {
	fs := flag.NewFlagSet("ping", flag.ExitOnError)
	bs := fs.String("beacon", "haz", "which beacon provider to use")
	fs.Parse(args)
	b, err := beacon.New(*bs)
	if err != nil {
		return "", fmt.Errorf("unable to create beacon: %v", err)
	}
	return b.Ping(ctx)
}
