package main

import (
	"flag"
	"fmt"
	"os"
)

// handleCheckCommand processes the 'check' subcommand, which detects and displays
// the current public IP address using a beacon service.
//
// The command accepts an optional -b flag to specify a particular beacon service.
// If no beacon is specified, it uses the configured default beacon or selects
// a random one from available options.
//
// The function will terminate the program with exit code 1 if:
//   - Invalid arguments are provided
//   - The beacon service cannot be contacted
//   - The IP address cannot be retrieved or validated
//
// On success, the detected IP address is printed to stdout.
func handleCheckCommand(cfg Config, args []string) {
	checkCmd := flag.NewFlagSet("check", flag.ExitOnError)
	beacon := checkCmd.String("b", "", "specify a beacon (optional)")

	if err := checkCmd.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: unable to parse subcommand arguments: %v\n", err)
		showUsage()
		os.Exit(1)
	}

	err := validateCmd(checkCmd.NArg())
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: unexpected arguments: %v: %v\n", checkCmd.Args(), err)
		showUsage()
		os.Exit(1)
	}

	beaconName := *beacon
	if beaconName == "" {
		beaconName = cfg.UserConfig.Beacon
	}

	_, err = contactBeacon(cfg, beaconName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: unable to get public IP from beacon %s: %v\n", beaconName, err)
		os.Exit(1)
	}
}

// handlePushCommand processes the 'push' subcommand, which detects the current
// public IP address and stores it in a tender service (like GitHub Gists).
//
// The command requires a tender service to be specified either via the -t flag
// or in the configuration file. An optional -b flag can specify which beacon
// service to use for IP detection.
//
// The function will terminate the program with exit code 1 if:
//   - Invalid arguments are provided
//   - No tender service is specified
//   - The beacon service fails to provide an IP address
//   - The tender service cannot be set up or accessed
//   - The push operation fails
//
// On success, the IP address is stored in the tender service and may be
// retrieved later using the pull command.
func handlePushCommand(cfg Config, args []string) {
	pushCmd := flag.NewFlagSet("push", flag.ExitOnError)
	beacon := pushCmd.String("b", "", "specify a beacon (optional)")
	tender := pushCmd.String("t", "", "specify a tender")

	if err := pushCmd.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: unable to parse subcommand arguments: %v\n", err)
		showUsage()
		os.Exit(1)
	}

	err := validateCmd(pushCmd.NArg())
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: unexpected arguments: %v: %v\n", pushCmd.Args(), err)
		showUsage()
		os.Exit(1)
	}

	tenderName := *tender
	if tenderName == "" {
		tenderName = cfg.UserConfig.Tender
	}
	if tenderName == "" {
		fmt.Fprintf(os.Stderr, "ERROR: tender must be specified with -t flag or configured in the configuration file\n")
		os.Exit(1)
	}

	beaconName := *beacon
	if beaconName == "" {
		beaconName = cfg.UserConfig.Beacon
	}

	publicIP, err := contactBeacon(cfg, beaconName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: unable to get public IP from beacon %s: %v\n", beaconName, err)
		os.Exit(1)
	}

	selectedTender, err := setupTender(cfg, tenderName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: unable to setup tender %s: %v\n", tenderName, err)
		os.Exit(1)
	}

	cfg, err = configLoad()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: unable to reload configuration file: %v\n", err)
		os.Exit(1)
	}

	_, err = pushTender(cfg, selectedTender, publicIP)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: unable to push public IP to tender %s: %v\n", tenderName, err)
		os.Exit(1)
	}
}

// handlePullCommand processes the 'pull' subcommand, which retrieves and displays
// stored IP addresses from a tender service.
//
// The command requires a tender service to be specified either via the -t flag
// or in the configuration file. It retrieves all IP addresses that have been
// previously stored using the push command.
//
// The function will terminate the program with exit code 1 if:
//   - Invalid arguments are provided
//   - No tender service is specified
//   - The tender service cannot be set up or accessed
//   - No stored data is found in the tender service
//   - The pull operation fails
//
// On success, all stored hostname-to-IP mappings are printed to stdout
// in the format "<hostname>:<address>".
func handlePullCommand(cfg Config, args []string) {
	pullCmd := flag.NewFlagSet("pull", flag.ExitOnError)
	tender := pullCmd.String("t", "", "specify a tender")

	if err := pullCmd.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: unable to parse subcommand arguments: %v\n", err)
		showUsage()
		os.Exit(1)
	}

	err := validateCmd(pullCmd.NArg())
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: unexpected arguments: %v: %v\n", pullCmd.Args(), err)
		showUsage()
		os.Exit(1)
	}

	tenderName := *tender
	if tenderName == "" {
		tenderName = cfg.UserConfig.Tender
	}
	if tenderName == "" {
		fmt.Fprintf(os.Stderr, "ERROR: tender must be specified with -t flag or configured in the configuration file\n")
		os.Exit(1)
	}

	selectedTender, err := setupTender(cfg, tenderName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: unable to setup tender %s: %v\n", tenderName, err)
		os.Exit(1)
	}

	cfg, err = configLoad()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: unable to reload configuration file: %v\n", err)
		os.Exit(1)
	}
	_, err = pullTender(cfg, selectedTender)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: unable to pull from tender %s: %v\n", tenderName, err)
		os.Exit(1)
	}
}
