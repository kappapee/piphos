package tender

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/kappapee/piphos/internal/config"
)

const (
	githubName = "gh"
	githubURL  = "https://api.github.com/gists"
)

// github implements the Tender interface using GitHub Gists as storage.
// It stores hostname-to-IP mappings in a private gist identified by the
// description "_piphos_" containing a single JSON file.
type github struct {
	baseURL string
	client  *http.Client
	headers map[string]string
	name    string
	token   string
}

// newGithub creates a GitHub tender with the provided authentication token.
func newGithub(token string) *github {
	return &github{
		baseURL: githubURL,
		client:  &http.Client{Timeout: config.HTTPClientTimeout},
		headers: map[string]string{
			"User-Agent":           config.PiphosUserAgent,
			"Accept":               "application/vnd.github+json",
			"X-GitHub-Api-Version": "2022-11-28",
		},
		name:  githubName,
		token: token,
	}
}

// gist represents the GitHub Gist API response structure.
type gist struct {
	ID          string              `json:"id"`
	Description string              `json:"description"`
	Public      bool                `json:"public"`
	Files       map[string]gistFile `json:"files"`
}

// gistFile represents a file within a GitHub Gist.
type gistFile struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

// Pull retrieves all hostname-to-IP mappings from the piphos GitHub Gist.
// Returns an error if the gist doesn't exist or cannot be parsed.
func (gh *github) Pull(ctx context.Context) (map[string]string, error) {
	gist, err := gh.readGist(ctx)
	if err != nil {
		return nil, err
	}
	if gist == nil {
		return nil, fmt.Errorf("no piphos gist found")
	}
	gistPiphosFile, ok := gist.Files[config.PiphosStamp]
	if !ok {
		return nil, fmt.Errorf("gist missing file: %s", config.PiphosStamp)
	}
	gistContentString := gistPiphosFile.Content
	var gistContent map[string]string
	if err := json.Unmarshal([]byte(gistContentString), &gistContent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal content: %w", err)
	}
	return gistContent, nil
}

// Push updates the IP address for the specified hostname in the GitHub Gist.
// If no piphos gist exists, a new private gist is created.
// If the hostname already has the same IP, no API call is made.
func (gh *github) Push(ctx context.Context, localHostname, publicIP string) error {
	gist, err := gh.readGist(ctx)
	if err != nil {
		return err
	}
	if gist == nil {
		return gh.createGist(ctx, localHostname, publicIP)
	}
	gistPiphosFile, ok := gist.Files[config.PiphosStamp]
	if !ok {
		return fmt.Errorf("gist missing file: %s", config.PiphosStamp)
	}
	gistContentString := gistPiphosFile.Content
	var gistContent map[string]string
	if err := json.Unmarshal([]byte(gistContentString), &gistContent); err != nil {
		return fmt.Errorf("failed to unmarshal content: %w", err)
	}
	// Skip update if IP hasn't changed
	if gistContent[localHostname] == publicIP {
		return nil
	}
	return gh.updateGist(ctx, gist.ID, gistContent, localHostname, publicIP)
}

// createGist creates a new private GitHub Gist with the initial hostname-to-IP mapping.
func (gh *github) createGist(ctx context.Context, localHostname, publicIP string) error {
	gistContent := map[string]string{localHostname: publicIP}
	content, err := json.Marshal(gistContent)
	if err != nil {
		return fmt.Errorf("failed to marshal content: %w", err)
	}
	gistPayload := gist{
		Description: config.PiphosStamp,
		Public:      false,
		Files: map[string]gistFile{
			config.PiphosStamp: {
				Filename: config.PiphosStamp,
				Content:  string(content),
			},
		},
	}
	gistRequestBody, err := json.Marshal(gistPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	if _, err := gh.gistRequest(ctx, http.MethodPost, gh.baseURL, http.StatusCreated, gistRequestBody); err != nil {
		return fmt.Errorf("failed to complete gist request: %w", err)
	}
	return nil
}

// readGist finds and retrieves the piphos gist.
// Returns nil if no piphos gist exists, which is not considered an error.
func (gh *github) readGist(ctx context.Context) (*gist, error) {
	gistsResponseBody, err := gh.gistRequest(ctx, http.MethodGet, gh.baseURL, http.StatusOK, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to complete gist request: %w", err)
	}
	var gists []gist
	if err := json.Unmarshal(gistsResponseBody, &gists); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	// Find the piphos gist by searching for the stamp description
	var gistID string
	for _, g := range gists {
		if g.Description == config.PiphosStamp {
			gistID = g.ID
			break
		}
	}
	// No piphos gist exists yet, not an error
	if gistID == "" {
		return nil, nil
	}
	URL := fmt.Sprintf("%s/%s", gh.baseURL, gistID)
	gistResponseBody, err := gh.gistRequest(ctx, http.MethodGet, URL, http.StatusOK, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to complete gist request: %w", err)
	}
	var gist gist
	if err := json.Unmarshal(gistResponseBody, &gist); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &gist, nil
}

// updateGist modifies an existing gist to update the hostname-to-IP mapping.
func (gh *github) updateGist(ctx context.Context, gistID string, gistContent map[string]string, localHostname, publicIP string) error {
	gistContent[localHostname] = publicIP
	content, err := json.Marshal(gistContent)
	if err != nil {
		return fmt.Errorf("failed to marshal content: %w", err)
	}
	gistPayload := gist{
		Description: config.PiphosStamp,
		Public:      false,
		Files: map[string]gistFile{
			config.PiphosStamp: {
				Filename: config.PiphosStamp,
				Content:  string(content),
			},
		},
	}
	gistRequestBody, err := json.Marshal(gistPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	URL := fmt.Sprintf("%s/%s", gh.baseURL, gistID)
	if _, err := gh.gistRequest(ctx, http.MethodPatch, URL, http.StatusOK, gistRequestBody); err != nil {
		return fmt.Errorf("failed to complete gist request: %w", err)
	}
	return nil
}

// gistRequest executes an HTTP request to the GitHub Gist API.
// It handles authentication, headers, and validates the response status code.
func (gh *github) gistRequest(ctx context.Context, HTTPMethod, URL string, expectedStatus int, requestBody []byte) ([]byte, error) {
	var requestBodyReader io.Reader
	if requestBody != nil {
		requestBodyReader = bytes.NewReader(requestBody)
	}
	req, err := http.NewRequestWithContext(ctx, HTTPMethod, URL, requestBodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	for k, v := range gh.headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+gh.token)
	resp, err := gh.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get response: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to close response body: %v\n", err)
		}
	}()
	if resp.StatusCode != expectedStatus {
		return nil, fmt.Errorf("unexpected response status: %d", resp.StatusCode)
	}
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return responseBody, nil
}
