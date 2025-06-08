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

	"github.com/kappapee/piphos/internal/config"
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

func contactBeacon(cfg config.Config, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("usage example: command <beaconName>")
	}
	if len(BeaconConfig) == 0 {
		return "", errors.New("no configured beacons found")
	}

	beacon := args[0]

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

	if resp.StatusCode == http.StatusOK {
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("unable to read response body from beacon %s: %v", selectedBeacon.Name, err)
			return "", err
		}
		publicIP := strings.TrimSpace(string(content))
		// TODO: validate IP address here (formatting)
		fmt.Printf("%s\n", publicIP)
		return publicIP, nil
	} else {
		log.Printf("expected response status '200 OK' from beacon %s, got: %d", selectedBeacon.Name, resp.StatusCode)
		return "", errors.New("beacon did not respond with status 200 OK")
	}
}
