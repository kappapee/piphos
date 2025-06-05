package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	if TenderToken == "" {
		log.Fatal("PIPHOS_TOKEN environment variable is required")
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	beacon, err := selectBeacon(BeaconHaz)
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

	tender, err := selectTender(TenderGithub)
	if err != nil {
		log.Printf("something went wrong trying to select a tender: %v", err)
		return
	}

	tender = loadTenderPayload(tender, publicIP, false)

	err = pushToTender(client, tender)
	if err != nil {
		log.Printf("unable to push to tender %s: %v\n", tender.Name, err)
		return
	}

	fmt.Println("DONE!")
}
