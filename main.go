package main

import (
	"fmt"
	"log"
)

var beacons []Beacon

func init() {
	beacons = []Beacon{
		{Name: "icanhazip", URL: "https://ipv4.icanhazip.com"},
		{Name: "aws", URL: "https://checkip.amazonaws.com"},
	}
}

func main() {
	beacon, err := selectBeacon(beacons)
	if err != nil {
		log.Printf("something went wrong trying to select a beacon: %v", err)
		return
	}
	publicIP, err := contactBeacon(beacon)
	if err != nil {
		log.Printf("something went wrong trying to contact beacon %s: %v\n", beacon.Name, err)
		return
	}
	fmt.Printf("beacon %s reported public IP: %s\n", beacon.Name, publicIP)
}
