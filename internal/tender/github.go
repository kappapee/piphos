package tender

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/kappapee/piphos/internal/config"
)

const (
	githubName = "gh"
	githubURL  = "https://api.github.com/gists"
)

type github struct {
	baseURL string
	client  *http.Client
	headers map[string]string
	name    string
	token   string
}

func newGithub() *github {
	return &github{
		baseURL: githubURL,
		client:  &http.Client{Timeout: config.HTTPClientTimeout},
		headers: map[string]string{
			"User-Agent":           config.PiphosUserAgent,
			"Accept":               "application/vnd.github+json",
			"X-GitHub-Api-Version": "2022-11-28",
		},
		name:  githubName,
		token: os.Getenv("PIPHOS_GITHUB_TOKEN"),
	}
}

type Gist struct {
	ID          string              `json:"id"`
	Description string              `json:"description"`
	Public      bool                `json:"public"`
	Files       map[string]gistFile `json:"files"`
}

type gistFile struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

func (gh *github) Pull(ctx context.Context) (map[string]string, error) {
	results := map[string]string{}
	gist, err := gh.fetchGistByDescription(ctx)
	if err != nil {
		return results, err
	}
	if gist == nil || len(gist.Files) == 0 {
		return results, nil
	}
	for _, f := range gist.Files {
		results[f.Filename] = f.Content
	}
	return map[string]string{}, nil
}

func (gh *github) Push(ctx context.Context, hostname, ip string) error {
	gist, err := gh.fetchGistByDescription(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch gist by description: %w", err)
	}
	if gist == nil {
		err := gh.createGist(ctx, hostname, ip)
		if err != nil {
			return err
		}
		return nil
	}
	for _, g := range gist.Files {
		if g.Filename == hostname && g.Content != ip {
			err := gh.updateGist(ctx, hostname, ip, gist.ID)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return nil
}

func (gh *github) fetchGistByDescription(ctx context.Context) (*Gist, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, gh.baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GET request for tender %s: %w", gh.name, err)
	}
	for k, v := range gh.headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+gh.token)
	resp, err := gh.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to receive GET response from tender %s: %w", gh.name, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to close GET response body: %v\n", err)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected GET response status from tender %s: %d", gh.name, resp.StatusCode)
	}
	var gists []Gist
	if err := json.NewDecoder(resp.Body).Decode(&gists); err != nil {
		return nil, fmt.Errorf("failed to decode GET response json from tender %s: %w", gh.name, err)
	}
	for _, g := range gists {
		if g.Description == config.PiphosStamp {
			return &g, nil
		}
	}
	return nil, nil
}

func (gh *github) createGist(ctx context.Context, hostname, ip string) error {
	initGistFile := gistFile{Filename: config.PiphosStamp, Content: config.PiphosStamp}
	file := gistFile{Filename: hostname, Content: ip}
	newGist := Gist{
		Description: config.PiphosStamp,
		Public:      false,
		Files: map[string]gistFile{
			config.PiphosStamp: initGistFile,
			hostname:           file,
		},
	}
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(newGist)
	if err != nil {
		return fmt.Errorf("failed to encode POST request json for tender %s: %w", gh.name, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, gh.baseURL, &buf)
	if err != nil {
		return fmt.Errorf("failed to create POST request for tender %s: %w", gh.name, err)
	}
	for k, v := range gh.headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+gh.token)
	resp, err := gh.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to receive POST response from tender %s: %w", gh.name, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to close POST response body: %v\n", err)
		}
	}()
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected POST response status from tender %s: %d", gh.name, resp.StatusCode)
	}
	return nil
}

func (gh *github) updateGist(ctx context.Context, hostname, ip, gistID string) error {
	file := gistFile{Filename: hostname, Content: ip}
	newGist := Gist{
		Description: config.PiphosStamp,
		Public:      false,
		Files:       map[string]gistFile{hostname: file},
	}
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(newGist)
	if err != nil {
		return fmt.Errorf("failed to encode PUT request json for tender %s: %w", gh.name, err)
	}
	patchURL := gh.baseURL + "/" + gistID
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, patchURL, &buf)
	if err != nil {
		return fmt.Errorf("failed to create PUT request for tender %s: %w", gh.name, err)
	}
	for k, v := range gh.headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+gh.token)
	resp, err := gh.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get PUT response from tender %s: %w", gh.name, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to close PUT response body: %v\n", err)
		}
	}()
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected PUT response status from tender %s: %d", gh.name, resp.StatusCode)
	}
	return nil
}
