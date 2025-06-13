package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Tender struct {
	Name    string            `json:"name"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Payload GithubGist        `json:"data"`
}

type GithubFile struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

type GithubGist struct {
	ID          string                `json:"id"`
	Description string                `json:"description"`
	Public      bool                  `json:"public"`
	Files       map[string]GithubFile `json:"files"`
}

const (
	TenderGithub       = "github"
	PayloadDescription = "piphos"
)

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

func setupTender(cfg Config, tender string) (Tender, error) {
	if len(TenderConfig) == 0 {
		return Tender{}, fmt.Errorf("no configured tenders found\n")
	}

	err := validateToken(cfg.UserConfig.Token, tender)
	if err != nil {
		return Tender{}, fmt.Errorf("token validation failed: %v\n", err)
	}

	var selectedTender Tender

	switch tender {
	case TenderGithub:
		selectedTender = TenderConfig[TenderGithub]
		selectedTender.Headers["Authorization"] += cfg.UserConfig.Token
		selectedTender.Payload.Description = PayloadDescription
		selectedTender.Payload.Public = false
	default:
		return Tender{}, fmt.Errorf("unknown tender requested\n")
	}

	if cfg.UserConfig.PiphosGistID == "" {
		req, err := http.NewRequest("GET", selectedTender.URL, nil)
		if err != nil {
			return Tender{}, fmt.Errorf("unable to create request for tender %s: %v\n", selectedTender.Name, err)
		}

		for k, v := range selectedTender.Headers {
			req.Header.Set(k, v)
		}

		resp, err := cfg.Client.Do(req)
		if err != nil {
			return Tender{}, fmt.Errorf("unable to get response from tender %s: %v\n", selectedTender.Name, err)
		}
		defer resp.Body.Close()

		respContent, err := io.ReadAll(resp.Body)
		if err != nil {
			return Tender{}, fmt.Errorf("unable to read response data from tender %s: %v\n", selectedTender.Name, err)
		}

		var gists []GithubGist
		err = json.Unmarshal(respContent, &gists)
		if err != nil {
			return Tender{}, fmt.Errorf("unable to unmarshal json data from tender %s: %v\n", selectedTender.Name, err)
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
		configSave(cfg)
	}
	return selectedTender, nil
}

func pushTender(cfg Config, tender Tender, ip string) (string, error) {
	tender.Payload.Files = map[string]GithubFile{
		cfg.UserConfig.Hostname: {
			Filename: cfg.UserConfig.Hostname,
			Content:  ip,
		},
	}

	jsonBody, err := json.Marshal(tender.Payload)
	if err != nil {
		return "", fmt.Errorf("unable to create json payload for tender %s: %v\n", tender.Name, err)
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
		return "", fmt.Errorf("unable to create request for tender %s: %v\n", tender.Name, err)
	}

	for k, v := range tender.Headers {
		req.Header.Set(k, v)
	}

	resp, err := cfg.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("unable to get response from tender %s: %v\n", tender.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("tender %s returned status %d\n", tender.Name, resp.StatusCode)
	}

	if cfg.UserConfig.PiphosGistID == "" {
		respContent, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("unable to read response data from tender %s: %v\n", tender.Name, err)
		}

		var gist GithubGist
		err = json.Unmarshal(respContent, &gist)
		if err != nil {
			return "", fmt.Errorf("unable to unmarshal json data from tender %s: %v\n", tender.Name, err)
		}

		cfg.UserConfig.PiphosGistID = gist.ID
		configSave(cfg)
	}
	return ip, nil
}

func pullTender(cfg Config, tender Tender) (string, error) {
	if cfg.UserConfig.PiphosGistID == "" {
		return "", fmt.Errorf("no piphos records on tender %s, try the push subcommand first\n", tender.Name)
	}

	url := tender.URL + "/" + cfg.UserConfig.PiphosGistID
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("unable to create request for tender %s: %v\n", tender.Name, err)
	}

	for k, v := range tender.Headers {
		req.Header.Set(k, v)
	}

	resp, err := cfg.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("unable to get response from tender %s: %v\n", tender.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("tender %s returned status %d\n", tender.Name, resp.StatusCode)
	}

	respContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("unable to read response data from tender %s: %v\n", tender.Name, err)
	}

	var gist GithubGist
	err = json.Unmarshal(respContent, &gist)
	if err != nil {
		return "", fmt.Errorf("unable to unmarshal json data from tender %s: %v\n", tender.Name, err)
	}
	if len(gist.Files) > 0 {
		for _, f := range gist.Files {
			fmt.Printf("%s:%s\n", f.Filename, f.Content)
		}
	}
	return "", nil
}
