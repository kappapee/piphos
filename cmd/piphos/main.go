package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kappapee/piphos/internal/command"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: piphos <command> (args)")
		os.Exit(1)
	}
	ctx := context.Background()
	switch os.Args[1] {
	case "ping":
		publicIP, err := command.Ping(ctx, os.Args[2:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to run ping command: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintln(os.Stdout, publicIP)
		os.Exit(0)
	}
}
