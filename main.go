package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kappapee/piphos/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("unable to load configuration file: %v\n", err)
		os.Exit(1)
	}

	if cfg.Token == "" {
		fmt.Println("TOKEN must be set in the configuration file")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Println("usage: piphos <command> [<args>]")
		os.Exit(1)
	}

	checkCmd := flag.NewFlagSet("check", flag.ExitOnError)
	var beacon string
	checkCmd.StringVar(&beacon, "b", "", "specify a beacon (optional)")

	pushCmd := flag.NewFlagSet("push", flag.ExitOnError)
	var tender string
	pushCmd.StringVar(&tender, "t", "", "specify a tender")

	switch os.Args[1] {
	case "check":
		checkCmd.Parse(os.Args[2:])
		_, err := contactBeacon(cfg, beacon)
		if err != nil {
			fmt.Printf("unable to get public IP from beacon %s: %v\n", beacon, err)
			os.Exit(1)
		}
	case "push":
		pushCmd.Parse(os.Args[2:])
		publicIP, err := contactBeacon(cfg, BeaconDefault)
		if err != nil {
			fmt.Printf("unable to get public IP: %v\n", err)
			os.Exit(1)
		}
		_, err = pushTender(cfg, tender, publicIP)
		if err != nil {
			fmt.Printf("unable to push public IP to tender %s: %v\n", tender, err)
			os.Exit(1)
		}
	default:
		fmt.Println("usage: piphos <command> [<args>]")
		os.Exit(1)
	}
}
