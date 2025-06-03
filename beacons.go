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

func selectBeacon(beacons []Beacon) (Beacon, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if len(beacons) == 0 {
		err := errors.New("beacons list is empty")
		return Beacon{}, err
	}
	return beacons[r.Intn(len(beacons))], nil
}

func contactBeacon(beacon Beacon) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

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
		err = errors.New("did not get 200 Status OK")
		return "", err
	}
}
