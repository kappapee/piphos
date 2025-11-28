package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kappapee/piphos/internal/exec"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: piphos <command> (args)")
		os.Exit(1)
	}
	ctx := context.Background()
	switch os.Args[1] {
	case "ping":
		publicIP, err := exec.Ping(ctx, os.Args[2:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to run ping command: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintln(os.Stdout, publicIP)
		os.Exit(0)
	case "pull":
		hosts, err := exec.Pull(ctx, os.Args[2:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to run pull command: %v\n", err)
			os.Exit(1)
		}
		for k, v := range hosts {
			fmt.Fprintf(os.Stdout, "%s: %s\n", k, v)
		}
		os.Exit(0)
	case "push":
		publicIP, err := exec.Ping(ctx, []string{"beacon"})
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to check public IP address: %v\n", err)
			os.Exit(1)
		}
		if err := exec.Push(ctx, os.Args[2:], publicIP); err != nil {
			fmt.Fprintf(os.Stderr, "failed to run push command: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintln(os.Stdout, publicIP)
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %v\n", os.Args[1])
	}
}
