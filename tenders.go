package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/kappapee/piphos/internal/config"
)

type Tender struct {
	Name    string            `json:"name"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Data    TenderPayload     `json:"data"`
}

type TenderPayload struct {
	Description string `json:"description"`
	Title       string `json:"title"`
	Public      bool   `json:"public"`
	Visibility  string `json:"visibility"`
	Files       any    `json:"files"`
}

const (
	TenderGithub = "github"
	TenderGitlab = "gitlab"
	TenderTitle  = "piphos"
)

var TenderConfig = map[string]Tender{
	TenderGithub: {
		Name: TenderGithub,
		URL:  "https://api.github.com/gists",
		Headers: map[string]string{
			"X-GitHub-Api-Version": "2022-11-28",
			"Content-Type":         "application/json",
		},
		Data: TenderPayload{
			Description: TenderTitle,
			Public:      false,
		},
	},
	TenderGitlab: {
		Name: TenderGitlab,
		URL:  "https://gitlab.com/api/v4/snippets",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Data: TenderPayload{
			Title:      TenderTitle,
			Visibility: "private",
		},
	},
}

func pushTender(cfg config.Config, args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("usage example: command <tenderName> <ip>")
	}
	if len(TenderConfig) == 0 {
		return "", errors.New("no configured tenders found")
	}

	tender := args[0]
	ip := args[1]

	var selectedTender Tender

	switch tender {
	case TenderGithub:
		selectedTender = TenderConfig[TenderGithub]
		selectedTender.Headers["Authorization"] = "Bearer " + cfg.Token
		selectedTender.Data.Files = map[string]map[string]string{cfg.Hostname: {"content": ip}}
	case TenderGitlab:
		selectedTender = TenderConfig[TenderGitlab]
		selectedTender.Headers["PRIVATE-TOKEN"] = cfg.Token
		selectedTender.Data.Files = []map[string]string{{"file_path": cfg.Hostname, "content": ip}}
	default:
		return "", errors.New("unknown tender requested")
	}

	jsonBody, err := json.Marshal(selectedTender.Data)
	if err != nil {
		log.Printf("unable to create json payload for tender %s: %v", selectedTender.Name, err)
		return "", err
	}
	bodyReader := bytes.NewReader(jsonBody)

	req, err := http.NewRequest("POST", selectedTender.URL, bodyReader)
	if err != nil {
		log.Printf("unable to create request for tender %s: %v", selectedTender.Name, err)
		return "", err
	}

	for k, v := range selectedTender.Headers {
		req.Header.Set(k, v)
	}

	resp, err := cfg.Client.Do(req)
	if err != nil {
		log.Printf("unable to get response from tender %s: %v", selectedTender.Name, err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("tender %s returned status %d", selectedTender.Name, resp.StatusCode)
		return "", fmt.Errorf("tender returned status %d", resp.StatusCode)
	}
	return ip, nil
}
