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
	id, err := gh.fetchGistIDByDescription(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch gist ID by description: %w", err)
	}
	if id == "" {
		return nil, fmt.Errorf("no hosts gist found on tender %s: %w", gh.name, err)
	}
	gist, err := gh.readGist(ctx, id)
	if err != nil {
		return nil, err
	}
	gistContent := gist.Files[config.PiphosStamp].Content
	var hosts map[string]string
	if err := json.Unmarshal([]byte(gistContent), &hosts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal hosts json from tender %s: %w", gh.name, err)
	}
	return hosts, nil
}

func (gh *github) Push(ctx context.Context, hostname, ip string) error {
	id, err := gh.fetchGistIDByDescription(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch gist ID by description: %w", err)
	}
	if id == "" {
		err := gh.createGist(ctx, hostname, ip)
		if err != nil {
			return err
		}
		return nil
	} else {
		gist, err := gh.readGist(ctx, id)
		if err != nil {
			return err
		}
		gistContent := gist.Files[config.PiphosStamp].Content
		var hosts map[string]string
		if err := json.Unmarshal([]byte(gistContent), &hosts); err != nil {
			return fmt.Errorf("failed to unmarshal hosts json from tender %s: %w", gh.name, err)
		}
		if hosts[hostname] != ip {
			err := gh.updateGist(ctx, id, hosts, hostname, ip)
			if err != nil {
				return err
			}
			return nil
		}
		return nil
	}
}

func (gh *github) fetchGistIDByDescription(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, gh.baseURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create GET request for tender %s: %w", gh.name, err)
	}
	for k, v := range gh.headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+gh.token)
	resp, err := gh.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to receive GET response from tender %s: %w", gh.name, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to close GET response body: %v\n", err)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected GET response status from tender %s: %d", gh.name, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read GET response body from tender %s: %w", gh.name, err)
	}
	var gists []Gist
	if err := json.Unmarshal(body, &gists); err != nil {
		return "", fmt.Errorf("failed to unmarshal GET response json from tender %s: %w", gh.name, err)
	}
	for _, g := range gists {
		if g.Description == config.PiphosStamp {
			return g.ID, nil
		}
	}
	return "", nil
}

func (gh *github) createGist(ctx context.Context, hostname, ip string) error {
	host := map[string]string{hostname: ip}
	hb, err := json.Marshal(host)
	if err != nil {
		return fmt.Errorf("failed to marshal host json for tender %s: %w", gh.name, err)
	}
	gf := gistFile{Filename: config.PiphosStamp, Content: string(hb)}
	gist := Gist{
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
	if _, err := gh.gistRequest(ctx, http.MethodPost, gh.baseURL, http.StatusCreated, bytes.NewReader(gb)); err != nil {
		return fmt.Errorf("failed to complete gist request: %w", err)
	}
	return nil
}

func (gh *github) readGist(ctx context.Context, id string) (*Gist, error) {
	var gist Gist
	url := gh.baseURL + "/" + id
	body, err := gh.gistRequest(ctx, http.MethodGet, url, http.StatusOK, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to complete gist request: %w", err)
	}
	if err := json.Unmarshal(body, &gist); err != nil {
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
	gist := Gist{
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
	url := gh.baseURL + "/" + id
	if _, err := gh.gistRequest(ctx, http.MethodPatch, url, http.StatusOK, bytes.NewReader(gb)); err != nil {
		return fmt.Errorf("failed to complete gist request: %w", err)
	}
	return nil
}

func (gh *github) gistRequest(ctx context.Context, method, url string, expectStatus int, body *bytes.Reader) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
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
	fmt.Fprintf(os.Stdout, "DEBUG: respBody: %v", respBody)
	return respBody, nil
}
