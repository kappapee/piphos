// Package main implements piphos, a command-line tool for managing dynamic IP addresses
// in homelabs. It provides functionality to detect public IP addresses using beacon
// services and store/retrieve them using tender services like GitHub Gists.
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
	"log"
	"os"
)

// main is the entry point for the piphos CLI application.
// It loads configuration, parses command-line arguments, and routes
// to the appropriate command handler.
func main() {
	cfg, err := configLoad()
	if err != nil {
		log.Fatalf("error: unable to load configuration file: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		log.Printf("error: not enough arguments\n")
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
		log.Printf("error: sub-command not found\n")
		showUsage()
		os.Exit(1)
	}
}
