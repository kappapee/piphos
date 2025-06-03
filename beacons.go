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
	Name string
	URL  string
}

func selectBeacon(beacon string) (Beacon, error) {
	beacons := map[string]Beacon{
		"haz": {Name: "icanhazip", URL: "https://ipv4.icanhazip.com"},
		"aws": {Name: "aws", URL: "https://checkip.amazonaws.com"},
	}

	if len(beacons) == 0 {
		return Beacon{}, errors.New("beacons list is empty")
	}

	switch beacon {
	case "aws":
		return beacons["aws"], nil
	case "haz":
		return beacons["haz"], nil
	default:
		mapKeys := make([]string, 0, len(beacons))
		for key := range beacons {
			mapKeys = append(mapKeys, key)
		}
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		return beacons[mapKeys[r.Intn(len(mapKeys))]], nil

	}
}

func contactBeacon(client *http.Client, beacon Beacon) (string, error) {
	req, err := http.NewRequest("GET", beacon.URL, nil)
	if err != nil {
		log.Printf("unable to create request to beacon %s: %v", beacon.Name, err)
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
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("unable to read response body from beacon %s: %v", beacon.Name, err)
			return "", err
		}
		bodyString := strings.TrimSpace(string(bodyBytes))
		return bodyString, nil
	} else {
		log.Printf("expected 200 Status OK from beacon %s, got: %d", beacon.Name, resp.StatusCode)
		return "", errors.New("did not get 200 Status OK")
	}
}
