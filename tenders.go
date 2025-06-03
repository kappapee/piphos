package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
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

func selectTender(tender string) (Tender, error) {
	tenderToken := os.Getenv("PIPHOS_TOKEN")
	if tenderToken == "" {
		return Tender{}, errors.New("tender token is not set")
	}

	tenders := map[string]Tender{
		"github": {
			Name: "GitHub",
			URL:  "https://api.github.com/gists",
			Headers: map[string]string{
				"Authorization":        "Bearer " + tenderToken,
				"X-GitHub-Api-Version": "2022-11-28",
			},
		},
		"gitlab": {
			Name: "GitLab",
			URL:  "https://gitlab.com/api/v4/snippets",
			Headers: map[string]string{
				"PRIVATE-TOKEN": tenderToken,
			},
		},
	}

	if len(tenders) == 0 {
		return Tender{}, errors.New("tenders list is empty")
	}

	switch tender {
	case "github":
		return tenders["github"], nil
	case "gitlab":
		return tenders["gitlab"], nil
	default:
		return Tender{}, errors.New("unknown tender requested")
	}
}

func loadTenderPayload(tender Tender, ip string, public bool) Tender {
	desc := "My server's public IP."
	filename := "piphos_by_pyculiar_labs"
	visibility := "private"
	if public {
		visibility = "public"
	}
	switch tender.Name {
	case "GitHub":
		tender.Data = TenderPayload{
			Description: desc,
			Public:      public,
			Files:       map[string]map[string]string{filename: {"content": ip}},
		}
	case "GitLab":
		tender.Data = TenderPayload{
			Title:      desc,
			Visibility: visibility,
			Files: []map[string]string{
				{"content": ip, "file_path": filename},
			},
		}
	default:
		tender.Data = TenderPayload{}
	}
	return tender
}

func pushToTender(client *http.Client, tender Tender) error {
	jsonBody, err := json.Marshal(tender.Data)
	if err != nil {
		log.Printf("unable to create json payload for tender %s: %v", tender.Name, err)
		return err
	}
	bodyReader := bytes.NewReader(jsonBody)

	req, err := http.NewRequest("POST", tender.URL, bodyReader)
	if err != nil {
		log.Printf("unable to create post request to beacon %s: %v", tender.Name, err)
		return err
	}

	for k, v := range tender.Headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("unable to get response from beacon %s: %v", tender.Name, err)
		return err
	}
	defer resp.Body.Close()
	return nil
}
