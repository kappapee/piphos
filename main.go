package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	beacons := []Beacon{
		{Name: "icanhazip", URL: "https://ipv4.icanhazip.com"},
		{Name: "aws", URL: "https://checkip.amazonaws.com"},
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	beacon, err := selectBeacon(beacons)
	if err != nil {
		log.Printf("something went wrong trying to select a beacon: %v", err)
		return
	}

	publicIP, err := contactBeacon(client, beacon)
	if err != nil {
		log.Printf("something went wrong trying to contact beacon %s: %v\n", beacon.Name, err)
		return
	}
	fmt.Printf("beacon %s reported public IP: %s\n", beacon.Name, publicIP)
}
