package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Tender represents a service that can store and retrieve IP address data.
// Tender services act as persistent storage backends, allowing users to
// store their current IP address and retrieve it later from other locations.
type Tender struct {
	// Name is a human-readable identifier for the tender service.
	Name string `json:"name"`

	// URL is the base API endpoint for the tender service.
	URL string `json:"url"`

	// Headers contains HTTP headers required for authentication and API versioning.
	Headers map[string]string `json:"headers"`

	// Payload contains the data structure used for API requests to this tender.
	Payload GithubGist `json:"data"`
}

// GithubFile represents a single file within a GitHub Gist.
// Each file contains a filename and its textual content.
type GithubFile struct {
	// Filename is the name of the file as it appears in the gist.
	Filename string `json:"filename"`

	// Content is the textual content of the file.
	// For piphos, this typically contains an IP address.
	Content string `json:"content"`
}

// GithubGist represents a GitHub Gist used for storing IP address data.
type GithubGist struct {
	// ID is the unique identifier assigned by GitHub to this gist.
	ID string `json:"id"`

	// Description is a human-readable description of the gist's purpose.
	Description string `json:"description"`

	// Public determines whether the gist is publicly visible or private.
	Public bool `json:"public"`

	// Files contains the files stored within this gist, keyed by filename.
	Files map[string]GithubFile `json:"files"`
}

// Tender service identifiers and configuration constants.
const (
	// TenderGithub identifies the GitHub Gists tender service.
	TenderGithub = "github"

	// PayloadDescription is the standard description used for piphos gists.
	// This helps identify gists created by piphos among other user gists.
	PayloadDescription = "_piphos_"
)

// TenderConfig maps tender identifiers to their corresponding Tender configurations.
// This registry contains all available tender services that can be used for
// IP address storage and retrieval. New tender services can be added by extending
// this map with additional entries.
//
// Each tender configuration includes the necessary API endpoints, headers,
// and payload structures required to interact with the service.
var TenderConfig = map[string]Tender{
	TenderGithub: {
		Name: TenderGithub,
		URL:  "https://api.github.com/gists",
		Headers: map[string]string{
			"X-GitHub-Api-Version": "2022-11-28",
			"Content-Type":         "application/json",
			"Authorization":        "Bearer ",
		},
		Payload: GithubGist{},
	},
}

// setupTender initializes and configures a tender service for use.
// It validates the authentication token, sets up the service configuration,
// and discovers any existing piphos gists to avoid creating duplicates.
//
// Parameters:
//   - cfg: Configuration containing authentication tokens and HTTP client
//   - tender: Identifier for the desired tender service
//
// Returns:
//   - Tender: Fully configured tender service ready for use
//   - error: An error if the service cannot be set up, authentication fails,
//     or if the tender identifier is unknown
//
// The function automatically saves any discovered gist IDs to the configuration
// file to avoid repeated API calls in future operations.
func setupTender(cfg Config, tender string) (Tender, error) {
	if len(TenderConfig) == 0 {
		return Tender{}, fmt.Errorf("no configured tenders found")
	}

	err := validateToken(cfg.UserConfig.Token, tender)
	if err != nil {
		return Tender{}, fmt.Errorf("token validation failed: %w", err)
	}

	var selectedTender Tender

	switch tender {
	case TenderGithub:
		selectedTender = TenderConfig[TenderGithub]
		selectedTender.Headers["Authorization"] += cfg.UserConfig.Token
		selectedTender.Payload.Description = PayloadDescription
		selectedTender.Payload.Public = false
	default:
		return Tender{}, fmt.Errorf("unknown tender requested")
	}

	if cfg.UserConfig.PiphosGistID == "" {
		req, err := http.NewRequest("GET", selectedTender.URL, nil)
		if err != nil {
			return Tender{}, fmt.Errorf("unable to create request for tender %s: %w", selectedTender.Name, err)
		}

		for k, v := range selectedTender.Headers {
			req.Header.Set(k, v)
		}

		resp, err := cfg.Client.Do(req)
		if err != nil {
			return Tender{}, fmt.Errorf("unable to get response from tender %s: %w", selectedTender.Name, err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "WARN: unable to close response body: %v\n", err)
			}
		}()

		respContent, err := io.ReadAll(resp.Body)
		if err != nil {
			return Tender{}, fmt.Errorf("unable to read response data from tender %s: %w", selectedTender.Name, err)
		}

		var gists []GithubGist
		err = json.Unmarshal(respContent, &gists)
		if err != nil {
			return Tender{}, fmt.Errorf("unable to unmarshal json data from tender %s: %w", selectedTender.Name, err)
		}

		gistID := ""
		if len(gists) > 0 {
			for _, gist := range gists {
				if gist.Description == PayloadDescription {
					gistID = gist.ID
				}
			}
		}

		cfg.UserConfig.PiphosGistID = gistID
		if err := configSave(cfg); err != nil {
			return Tender{}, fmt.Errorf("unable to save configuration file: %w", err)
		}
	}
	return selectedTender, nil
}

// pushTender stores an IP address in the specified tender service.
// If no existing gist is configured, it creates a new one. Otherwise,
// it updates the existing gist with the new IP address information.
//
// Parameters:
//   - cfg: Configuration containing authentication and gist ID information
//   - tender: Configured tender service to use for storage
//   - ip: IP address to store in the tender service
//
// Returns:
//   - string: The IP address that was successfully stored
//   - error: An error if the storage operation fails, authentication is rejected,
//     or if the tender service returns an error response
//
// The function automatically saves new gist IDs to the configuration file
// for future operations.
func pushTender(cfg Config, tender Tender, ip string) (string, error) {
	tender.Payload.Files = map[string]GithubFile{
		PayloadDescription: {
			Filename: PayloadDescription,
			Content:  PayloadDescription,
		},
		cfg.UserConfig.Hostname: {
			Filename: cfg.UserConfig.Hostname,
			Content:  ip,
		},
	}

	jsonBody, err := json.Marshal(tender.Payload)
	if err != nil {
		return "", fmt.Errorf("unable to create json payload for tender %s: %w", tender.Name, err)
	}
	bodyReader := bytes.NewReader(jsonBody)

	var url string
	var method string
	if cfg.UserConfig.PiphosGistID == "" {
		url = tender.URL
		method = "POST"
	} else {
		url = tender.URL + "/" + cfg.UserConfig.PiphosGistID
		method = "PATCH"
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return "", fmt.Errorf("unable to create request for tender %s: %w", tender.Name, err)
	}

	for k, v := range tender.Headers {
		req.Header.Set(k, v)
	}

	resp, err := cfg.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("unable to get response from tender %s: %w", tender.Name, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "WARN: unable to close response body: %v\n", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("tender %s returned status %d", tender.Name, resp.StatusCode)
	}

	if cfg.UserConfig.PiphosGistID == "" {
		respContent, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("unable to read response data from tender %s: %w", tender.Name, err)
		}

		var gist GithubGist
		err = json.Unmarshal(respContent, &gist)
		if err != nil {
			return "", fmt.Errorf("unable to unmarshal json data from tender %s: %w", tender.Name, err)
		}

		cfg.UserConfig.PiphosGistID = gist.ID
		if err := configSave(cfg); err != nil {
			return "", fmt.Errorf("unable to save configuration file: %w", err)
		}
	}

	return ip, nil
}

// pullTender retrieves stored IP address data from the specified tender service.
// It fetches the gist data and displays all stored hostname-to-IP mappings
// that have been previously saved using the push operation.
//
// Parameters:
//   - cfg: Configuration containing the gist ID and authentication information
//   - tender: Configured tender service to use for retrieval
//
// Returns:
//   - string: Empty string (display output is printed directly to stdout)
//   - error: An error if no gist ID is configured, the retrieval fails,
//     authentication is rejected, or if the response cannot be parsed
//
// The function prints each discovered hostname-to-IP mapping in the format
// "<hostname>:<address>" to stdout.
func pullTender(cfg Config, tender Tender) (string, error) {
	if cfg.UserConfig.PiphosGistID == "" {
		return "", fmt.Errorf("no piphos records on tender %s or piphos record ID not configured, try the push subcommand first", tender.Name)
	}

	url := tender.URL + "/" + cfg.UserConfig.PiphosGistID
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("unable to create request for tender %s: %w", tender.Name, err)
	}

	for k, v := range tender.Headers {
		req.Header.Set(k, v)
	}

	resp, err := cfg.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("unable to get response from tender %s: %w", tender.Name, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "WARN: unable to close response body: %v\n", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("tender %s returned status %d", tender.Name, resp.StatusCode)
	}

	respContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("unable to read response data from tender %s: %w", tender.Name, err)
	}

	var gist GithubGist
	err = json.Unmarshal(respContent, &gist)
	if err != nil {
		return "", fmt.Errorf("unable to unmarshal json data from tender %s: %w", tender.Name, err)
	}
	if len(gist.Files) > 0 {
		for _, f := range gist.Files {
			if f.Filename != PayloadDescription {
				fmt.Printf("%s:%s\n", f.Filename, f.Content)
			}
		}
	}

	return "", nil
}
