package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

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

type piphosData struct {
	Hostname string `json:"hostname"`
	PublicIP string `json:"public_ip"`
}

type GithubGistsResponse []struct {
	URL        string `json:"url"`
	ForksURL   string `json:"forks_url"`
	CommitsURL string `json:"commits_url"`
	ID         string `json:"id"`
	NodeID     string `json:"node_id"`
	GitPullURL string `json:"git_pull_url"`
	GitPushURL string `json:"git_push_url"`
	HTMLURL    string `json:"html_url"`
	Files      map[string]struct {
		Filename string `json:"filename"`
		Type     string `json:"type"`
		Language string `json:"language"`
		RawURL   string `json:"raw_url"`
		Size     int    `json:"size"`
	} `json:"files"`
	Public      bool      `json:"public"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Description string    `json:"description"`
	Comments    int       `json:"comments"`
	User        any       `json:"user"`
	CommentsURL string    `json:"comments_url"`
	Owner       struct {
		Login             string `json:"login"`
		ID                int    `json:"id"`
		NodeID            string `json:"node_id"`
		AvatarURL         string `json:"avatar_url"`
		GravatarID        string `json:"gravatar_id"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		OrganizationsURL  string `json:"organizations_url"`
		ReposURL          string `json:"repos_url"`
		EventsURL         string `json:"events_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		Type              string `json:"type"`
		SiteAdmin         bool   `json:"site_admin"`
	} `json:"owner"`
	Truncated bool `json:"truncated"`
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

func pushTender(cfg config.Config, tender, ip string) (string, error) {
	if len(TenderConfig) == 0 {
		return "", errors.New("no configured tenders found")
	}

	err := validateToken(cfg.Token, tender)
	if err != nil {
		return "", err
	}

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
		return "", fmt.Errorf("tender %s returned status %d", selectedTender.Name, resp.StatusCode)
	}
	return ip, nil
}

func pullTender(cfg config.Config, tender string) error {
	if len(TenderConfig) == 0 {
		return errors.New("no configured tenders found")
	}

	err := validateToken(cfg.Token, tender)
	if err != nil {
		return err
	}

	var selectedTender Tender

	switch tender {
	case TenderGithub:
		selectedTender = TenderConfig[TenderGithub]
		selectedTender.Headers["Authorization"] = "Bearer " + cfg.Token
	case TenderGitlab:
		selectedTender = TenderConfig[TenderGitlab]
		selectedTender.Headers["PRIVATE-TOKEN"] = cfg.Token
	default:
		return errors.New("unknown tender requested")
	}

	jsonBody, err := json.Marshal(selectedTender.Data)
	if err != nil {
		log.Printf("unable to create json payload for tender %s: %v", selectedTender.Name, err)
		return err
	}
	bodyReader := bytes.NewReader(jsonBody)

	req, err := http.NewRequest("GET", selectedTender.URL, nil)
	if err != nil {
		log.Printf("unable to create request for tender %s: %v", selectedTender.Name, err)
		return err
	}

	for k, v := range selectedTender.Headers {
		req.Header.Set(k, v)
	}

	resp, err := cfg.Client.Do(req)
	if err != nil {
		log.Printf("unable to get response from tender %s: %v", selectedTender.Name, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("tender %s returned status %d", selectedTender.Name, resp.StatusCode)
		return fmt.Errorf("tender %s returned status %d", selectedTender.Name, resp.StatusCode)
	}
}
