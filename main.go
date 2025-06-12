package main

import (
	"fmt"
	"os"
)

func main() {
	cfg, err := configLoad()
	if err != nil {
		fmt.Printf("unable to load configuration file: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
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
		showUsage()
		os.Exit(1)
	}
}
