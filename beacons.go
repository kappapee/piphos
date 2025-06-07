package main

import (
	"errors"
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
	BeaconHaz = "haz"
	BeaconAws = "aws"
)

var BeaconConfig = map[string]Beacon{
	BeaconHaz: {Name: "icanhazip", URL: "https://ipv4.icanhazip.com"},
	BeaconAws: {Name: "aws", URL: "https://checkip.amazonaws.com"},
}

func selectBeacon(beacon string) (Beacon, error) {
	if len(BeaconConfig) == 0 {
		return Beacon{}, errors.New("no configured beacons found")
	}

	switch beacon {
	case BeaconAws:
		return BeaconConfig[BeaconAws], nil
	case BeaconHaz:
		return BeaconConfig[BeaconHaz], nil
	default:
		keys := make([]string, 0, len(BeaconConfig))
		for k := range BeaconConfig {
			keys = append(keys, k)
		}
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		return BeaconConfig[keys[r.Intn(len(keys))]], nil
	}
}

func contactBeacon(client *http.Client, beacon Beacon) (string, error) {
	req, err := http.NewRequest("GET", beacon.URL, nil)
	if err != nil {
		log.Printf("unable to create request for beacon %s: %v", beacon.Name, err)
		return "", err
	}
	req.Header.Set("User-Agent", "piphos/0.1")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("unable to get response from beacon %s: %v", beacon.Name, err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("unable to read response body from beacon %s: %v", beacon.Name, err)
			return "", err
		}
		publicIP := strings.TrimSpace(string(content))
		// TODO: validate IP address here (formatting)
		return publicIP, nil
	} else {
		log.Printf("expected response status '200 OK' from beacon %s, got: %d", beacon.Name, resp.StatusCode)
		return "", errors.New("beacon did not respond with status 200 OK")
	}
}
