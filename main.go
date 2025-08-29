// Package main implements piphos, a command-line tool for managing dynamic IP addresses
// in homelabs. It provides functionality to detect public IP addresses using beacon
// services and store/retrieve them using tender services.
//
// The tool supports three main commands:
//   - check: Detect and display the current public IP address
//   - push: Store the current public IP address to a tender service
//   - pull: Retrieve stored IP addresses from a tender service
//
// Configuration is managed through a JSON file stored in the user's configuration
// directory, with support for multiple beacon and tender providers.
package main

import (
	"fmt"
	"os"
)

// main is the entry point for the piphos CLI application.
// It loads configuration, parses command-line arguments, and routes
// to the appropriate command handler.
func main() {
	cfg, err := configLoad()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: unable to load configuration file: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, "ERROR: not enough arguments provided\n")
		showUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "check":
		handleCheckCommand(cfg, os.Args[2:])
	case "push":
		handlePushCommand(cfg, os.Args[2:])
	case "pull":
		handlePullCommand(cfg, os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "ERROR: subcommand %v not found\n", os.Args[1])
		showUsage()
		os.Exit(1)
	}
}
