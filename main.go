package main

import (
	"fmt"
	"os"
)

func main() {
	cli := NewCLI()

	cli.AddCommand("check", "check public IP using a beacon", contactBeacon)
	cli.AddCommand("push", "push public IP to tender", pushTender)

	if _, err := cli.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
