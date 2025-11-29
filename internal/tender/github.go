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

type gist struct {
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
	gist, err := gh.readGist(ctx)
	if err != nil {
		return nil, err
	}
	if gist == nil {
		return nil, fmt.Errorf("no hosts gist found on tender %s", gh.name)
	}
	gf, ok := gist.Files[config.PiphosStamp]
	if !ok {
		return nil, fmt.Errorf("gist missing %s file on tender %s", config.PiphosStamp, gh.name)
	}
	gistContent := gf.Content
	var hosts map[string]string
	if err := json.Unmarshal([]byte(gistContent), &hosts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal hosts json from tender %s: %w", gh.name, err)
	}
	return hosts, nil
}

func (gh *github) Push(ctx context.Context, hostname, ip string) error {
	gist, err := gh.readGist(ctx)
	if err != nil {
		return err
	}
	if gist == nil {
		return gh.createGist(ctx, hostname, ip)
	}
	gf, ok := gist.Files[config.PiphosStamp]
	if !ok {
		return fmt.Errorf("gist missing %s file on tender %s", config.PiphosStamp, gh.name)
	}
	gistContent := gf.Content
	var hosts map[string]string
	if err := json.Unmarshal([]byte(gistContent), &hosts); err != nil {
		return fmt.Errorf("failed to unmarshal hosts json from tender %s: %w", gh.name, err)
	}
	if hosts[hostname] != ip {
		return gh.updateGist(ctx, gist.ID, hosts, hostname, ip)
	}
	return nil
}

func (gh *github) createGist(ctx context.Context, hostname, ip string) error {
	host := map[string]string{hostname: ip}
	hb, err := json.Marshal(host)
	if err != nil {
		return fmt.Errorf("failed to marshal host json for tender %s: %w", gh.name, err)
	}
	gf := gistFile{Filename: config.PiphosStamp, Content: string(hb)}
	gist := gist{
		Description: config.PiphosStamp,
		Public:      false,
		Files: map[string]gistFile{
			config.PiphosStamp: gf,
		},
	}
	gb, err := json.Marshal(gist)
	if err != nil {
		return fmt.Errorf("failed to marshal POST request json for tender %s: %w", gh.name, err)
	}
	if _, err := gh.gistRequest(ctx, http.MethodPost, gh.baseURL, http.StatusCreated, gb); err != nil {
		return fmt.Errorf("failed to complete gist request: %w", err)
	}
	return nil
}

func (gh *github) readGist(ctx context.Context) (*gist, error) {
	gistsBody, err := gh.gistRequest(ctx, http.MethodGet, gh.baseURL, http.StatusOK, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to complete gist request: %w", err)
	}
	var gists []gist
	if err := json.Unmarshal(gistsBody, &gists); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response from tender %s: %w", gh.name, err)
	}
	var gistID string
	for _, g := range gists {
		if g.Description == config.PiphosStamp {
			gistID = g.ID
			break
		}
	}
	if gistID == "" {
		return nil, nil
	}
	url := fmt.Sprintf("%s/%s", gh.baseURL, gistID)
	gistBody, err := gh.gistRequest(ctx, http.MethodGet, url, http.StatusOK, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to complete gist request: %w", err)
	}
	var gist gist
	if err := json.Unmarshal(gistBody, &gist); err != nil {
		return nil, fmt.Errorf("failed to unmarshal GET response json from tender %s: %w", gh.name, err)
	}
	return &gist, nil
}

func (gh *github) updateGist(ctx context.Context, id string, hosts map[string]string, hostname, ip string) error {
	hosts[hostname] = ip
	hb, err := json.Marshal(hosts)
	if err != nil {
		return fmt.Errorf("failed to marshal host json for tender %s: %w", gh.name, err)
	}
	gf := gistFile{Filename: config.PiphosStamp, Content: string(hb)}
	gist := gist{
		Description: config.PiphosStamp,
		Public:      false,
		Files: map[string]gistFile{
			config.PiphosStamp: gf,
		},
	}
	gb, err := json.Marshal(gist)
	if err != nil {
		return fmt.Errorf("failed to marshal PATCH request json for tender %s: %w", gh.name, err)
	}
	url := fmt.Sprintf("%s/%s", gh.baseURL, id)
	if _, err := gh.gistRequest(ctx, http.MethodPatch, url, http.StatusOK, gb); err != nil {
		return fmt.Errorf("failed to complete gist request: %w", err)
	}
	return nil
}

func (gh *github) gistRequest(ctx context.Context, method, url string, expectStatus int, body []byte) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for tender %s: %w", gh.name, err)
	}
	for k, v := range gh.headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+gh.token)
	resp, err := gh.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get response from tender %s: %w", gh.name, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to close response body: %v\n", err)
		}
	}()
	if resp.StatusCode != expectStatus {
		return nil, fmt.Errorf("unexpected response status from tender %s: %d", gh.name, resp.StatusCode)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from tender %s: %w", gh.name, err)
	}
	return respBody, nil
}
