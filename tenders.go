package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
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
	if TenderToken == "" {
		return Tender{}, errors.New("tender token is not set")
	}

	if len(TenderConfig) == 0 {
		return Tender{}, errors.New("tenders list is empty")
	}

	var selectedTender Tender
	switch tender {
	case TenderGithub:
		selectedTender = TenderConfig[TenderGithub]
		selectedTender.Headers["Authorization"] = "Bearer " + TenderToken
	case TenderGitlab:
		selectedTender = TenderConfig[TenderGitlab]
		selectedTender.Headers["PRIVATE-TOKEN"] = TenderToken
	default:
		return Tender{}, errors.New("unknown tender requested")
	}
	return selectedTender, nil
}

func loadTenderPayload(tender Tender, ip string, public bool) Tender {
	desc := "My server's public IP."
	filename := "piphos_by_pyculiar_labs"
	visibility := "private"
	if public {
		visibility = "public"
	}
	switch tender.Name {
	case TenderGithub:
		tender.Data = TenderPayload{
			Description: desc,
			Public:      public,
			Files:       map[string]map[string]string{filename: {"content": ip}},
		}
	case TenderGitlab:
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
		log.Printf("unable to create post request to tender %s: %v", tender.Name, err)
		return err
	}

	for k, v := range tender.Headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("unable to get response from tender %s: %v", tender.Name, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("tender %s returned status %d", tender.Name, resp.StatusCode)
		return fmt.Errorf("tender returned status %d", resp.StatusCode)
	}
	return nil
}
