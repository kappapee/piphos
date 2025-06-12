package main

import (
	"flag"
	"fmt"
	"os"
)

func handleCheckCommand(cfg Config, args []string) {
	checkCmd := flag.NewFlagSet("check", flag.ExitOnError)
	beacon := checkCmd.String("b", "", "specify a beacon (optional)")

	checkCmd.Parse(args)

	if checkCmd.NArg() > 0 {
		fmt.Printf("error: unexpected arguments: %v\n", checkCmd.Args())
		os.Exit(1)
	}

	beaconName := *beacon
	if beaconName == "" {
		beaconName = cfg.UserConfig.Beacon
	}

	_, err := contactBeacon(cfg, beaconName)
	if err != nil {
		fmt.Printf("error: unable to get public IP from beacon %s: %v\n", beaconName, err)
		os.Exit(1)
	}
}

func handlePushCommand(cfg Config, args []string) {
	pushCmd := flag.NewFlagSet("push", flag.ExitOnError)
	beacon := pushCmd.String("b", "", "specify a beacon (optional)")
	tender := pushCmd.String("t", "", "specify a tender")

	pushCmd.Parse(args)

	if pushCmd.NArg() > 0 {
		fmt.Printf("error: unexpected arguments: %v\n", pushCmd.Args())
		os.Exit(1)
	}

	tenderName := *tender
	if tenderName == "" {
		tenderName = cfg.UserConfig.Tender
	}
	if tenderName == "" {
		fmt.Printf("error: tender must be specified with -t flag or configured in config file\n")
		os.Exit(1)
	}

	beaconName := *beacon
	if beaconName == "" {
		beaconName = cfg.UserConfig.Beacon
	}

	publicIP, err := contactBeacon(cfg, beaconName)
	if err != nil {
		fmt.Printf("error: unable to get public IP from beacon %s: %v\n", beaconName, err)
		os.Exit(1)
	}

	selectedTender, err := setupTender(cfg, tenderName)
	if err != nil {
		fmt.Printf("error: unable to setup tender %s: %v\n", tenderName, err)
		os.Exit(1)
	}

	_, err = pushTender(cfg, selectedTender, publicIP)
	if err != nil {
		fmt.Printf("error: unable to push public IP to tender %s: %v\n", tenderName, err)
		os.Exit(1)
	}
}

func handlePullCommand(cfg Config, args []string) {
	pullCmd := flag.NewFlagSet("pull", flag.ExitOnError)
	tender := pullCmd.String("t", "", "specify a tender")

	pullCmd.Parse(args)

	if pullCmd.NArg() > 0 {
		fmt.Printf("error: unexpected arguments: %v\n", pullCmd.Args())
		os.Exit(1)
	}

	tenderName := *tender
	if tenderName == "" {
		tenderName = cfg.UserConfig.Tender
	}
	if tenderName == "" {
		fmt.Printf("error: tender must be specified with -t flag or configured in config file\n")
		os.Exit(1)
	}

	selectedTender, err := setupTender(cfg, tenderName)
	if err != nil {
		fmt.Printf("error: unable to setup tender %s: %v\n", tenderName, err)
		os.Exit(1)
	}

	_, err = pullTender(cfg, selectedTender)
	if err != nil {
		fmt.Printf("error: unable to pull from tender %s: %v\n", tenderName, err)
		os.Exit(1)
	}
}
