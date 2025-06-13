package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// Beacon represents a service that provides public IP address detection.
// Beacon services are external HTTP endpoints that return the caller's
// public IP address in plain text format.
type Beacon struct {
	// Name is a human-readable identifier for the beacon service.
	Name string `json:"name"`

	// URL is the HTTP endpoint that returns the public IP address.
	// The endpoint should return the IP address as plain text.
	URL string `json:"url"`
}

// Beacon service identifiers used to select specific beacon providers.
const (
	BeaconDefault = ""
	BeaconHaz     = "haz"
	BeaconAws     = "aws"
)

// BeaconConfig maps beacon identifiers to their corresponding Beacon configurations.
// This registry contains all available beacon services that can be used for
// public IP address detection. New beacon services can be added by extending
// this map with additional entries.
var BeaconConfig = map[string]Beacon{
	BeaconHaz: {Name: "icanhazip", URL: "https://ipv4.icanhazip.com"},
	BeaconAws: {Name: "aws", URL: "https://checkip.amazonaws.com"},
}

// contactBeacon detects the current public IP address using the specified beacon service.
// If the beacon parameter is empty or refers to an unknown service, a random beacon
// is selected from the available configured options.
//
// Parameters:
//   - cfg: Configuration containing the HTTP client and other settings
//   - beacon: Identifier for the desired beacon service (empty string for random selection)
//
// Returns:
//   - string: The detected public IP address in string format
//   - error: An error if the beacon cannot be contacted, returns invalid data,
//     or if no beacon services are configured
func contactBeacon(cfg Config, beacon string) (string, error) {
	if len(BeaconConfig) == 0 {
		return "", fmt.Errorf("no configured beacons found\n")
	}

	var selectedBeacon Beacon

	switch beacon {
	case BeaconAws:
		selectedBeacon = BeaconConfig[BeaconAws]
	case BeaconHaz:
		selectedBeacon = BeaconConfig[BeaconHaz]
	default:
		keys := make([]string, 0, len(BeaconConfig))
		for k := range BeaconConfig {
			keys = append(keys, k)
		}
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		selectedBeacon = BeaconConfig[keys[r.Intn(len(keys))]]
		log.Printf("info: no beacon or unknown beacon provided, selecting random beacon: %s\n", selectedBeacon.Name)
	}

	req, err := http.NewRequest("GET", selectedBeacon.URL, nil)
	if err != nil {
		return "", fmt.Errorf("unable to create request for beacon %s: %w\n", selectedBeacon.Name, err)
	}
	req.Header.Set("User-Agent", "piphos/0.1")

	resp, err := cfg.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("unable to get response from beacon %s: %w\n", selectedBeacon.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("beacon %s returned status %d: %s\n", selectedBeacon.Name, resp.StatusCode, string(body))
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("unable to read response body from beacon %s: %w\n", selectedBeacon.Name, err)
	}

	publicIP := strings.TrimSpace(string(content))
	err = validateIP(publicIP)
	if err != nil {
		return "", err
	}

	fmt.Printf("%s\n", publicIP)
	return publicIP, nil
}
