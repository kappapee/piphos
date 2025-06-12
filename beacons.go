package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type Beacon struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

const (
	BeaconDefault = ""
	BeaconHaz     = "haz"
	BeaconAws     = "aws"
)

var BeaconConfig = map[string]Beacon{
	BeaconHaz: {Name: "icanhazip", URL: "https://ipv4.icanhazip.com"},
	BeaconAws: {Name: "aws", URL: "https://checkip.amazonaws.com"},
}

func contactBeacon(cfg Config, beacon string) (string, error) {
	if len(BeaconConfig) == 0 {
		return "", errors.New("no configured beacons found")
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
		if beacon == "" {
			fmt.Printf("selecting random beacon: %s\n", selectedBeacon.URL)
		} else {
			fmt.Printf("unknown beacon provided, selecting random beacon: %s\n", selectedBeacon.URL)
		}
	}

	req, err := http.NewRequest("GET", selectedBeacon.URL, nil)
	if err != nil {
		log.Printf("unable to create request for beacon %s: %v", selectedBeacon.Name, err)
		return "", err
	}
	req.Header.Set("User-Agent", "piphos/0.1")

	resp, err := cfg.Client.Do(req)
	if err != nil {
		log.Printf("unable to get response from beacon %s: %v", selectedBeacon.Name, err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("beacon %s returned status %d: %s", selectedBeacon.Name, resp.StatusCode, string(body))
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("unable to read response body from beacon %s: %v", selectedBeacon.Name, err)
		return "", err
	}

	publicIP := strings.TrimSpace(string(content))
	err = validateIP(publicIP)
	if err != nil {
		return "", err
	}

	fmt.Printf("%s\n", publicIP)
	return publicIP, nil
}
