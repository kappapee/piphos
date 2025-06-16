package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func NewGistResponse(description string, files map[string]GithubFile) string {
	return JSONResponse(GithubGist{
		Description: description,
		Files:       files,
	})
}

func TestSetupTender(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		tender  string
		status  int
		body    string
		wantErr bool
	}{
		{
			name:    "successful setup",
			cfg:     Config{UserConfig: UserConfig{Token: ValidClassicPAT}},
			tender:  TenderGithub,
			status:  200,
			body:    JSONResponse([]GithubGist{{ID: TestGistID, Description: PayloadDescription}}),
			wantErr: false,
		},
		{
			name:    "successful setup with fine-grained PAT",
			cfg:     Config{UserConfig: UserConfig{Token: ValidFineGrainedPAT}},
			tender:  TenderGithub,
			status:  200,
			body:    JSONResponse([]GithubGist{{ID: TestGistID, Description: PayloadDescription}}),
			wantErr: false,
		},
		{
			name:    "invalid token",
			cfg:     Config{UserConfig: UserConfig{Token: "invalid_token"}},
			tender:  TenderGithub,
			status:  200,
			body:    "",
			wantErr: true,
		},
		{
			name:    "unknown tender",
			cfg:     Config{UserConfig: UserConfig{Token: ValidClassicPAT}},
			tender:  "unknown",
			status:  200,
			body:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := NewMockTransport()
			transport.SetResponse(tt.status, tt.body)
			tt.cfg.Client = &http.Client{Transport: transport}

			tender, err := setupTender(tt.cfg, tt.tender)
			if (err != nil) != tt.wantErr {
				t.Errorf("setupTender() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tender.Name != tt.tender {
					t.Errorf("setupTender() tender name = %v, want %v", tender.Name, tt.tender)
				}
				if !strings.HasPrefix(tender.Headers["Authorization"], "Bearer "+tt.cfg.UserConfig.Token) {
					t.Error("Token not properly set in authorization header")
				}
				req := transport.LastRequest()
				if req != nil {
					VerifyRequest(t, req, "GET", "https://api.github.com/gists")
				}
			}
		})
	}
}

func TestPushTender(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		tender  Tender
		ip      string
		status  int
		body    string
		wantErr bool
	}{
		{
			name: "successful push new gist",
			cfg: Config{
				UserConfig: UserConfig{
					Hostname: TestHostname,
				},
			},
			tender: Tender{
				Name: TenderGithub,
				URL:  "https://api.github.com/gists",
				Headers: map[string]string{
					"Authorization": "Bearer " + ValidClassicPAT,
				},
			},
			ip:      TestIP,
			status:  201,
			body:    NewGistResponse("", map[string]GithubFile{TestHostname: {Filename: TestHostname, Content: TestIP}}),
			wantErr: false,
		},
		{
			name: "successful push existing gist",
			cfg: Config{
				UserConfig: UserConfig{
					Hostname:     TestHostname,
					PiphosGistID: TestGistID,
				},
			},
			tender: Tender{
				Name: TenderGithub,
				URL:  "https://api.github.com/gists",
				Headers: map[string]string{
					"Authorization": "Bearer " + ValidClassicPAT,
				},
			},
			ip:      TestIP,
			status:  200,
			body:    "{}",
			wantErr: false,
		},
		{
			name: "server error",
			cfg: Config{
				UserConfig: UserConfig{
					Hostname: TestHostname,
				},
			},
			tender: Tender{
				Name: TenderGithub,
				URL:  "https://api.github.com/gists",
				Headers: map[string]string{
					"Authorization": "Bearer " + ValidClassicPAT,
				},
			},
			ip:      TestIP,
			status:  401,
			body:    "Unauthorized",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := NewMockTransport()
			url := tt.tender.URL
			if tt.cfg.UserConfig.PiphosGistID != "" {
				url += "/" + tt.cfg.UserConfig.PiphosGistID
			}
			transport.SetResponse(tt.status, tt.body)
			tt.cfg.Client = &http.Client{Transport: transport}

			gotIP, err := pushTender(tt.cfg, tt.tender, tt.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("pushTender() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if gotIP != tt.ip {
					t.Errorf("pushTender() = %v, want %v", gotIP, tt.ip)
				}
				req := transport.LastRequest()
				if req == nil {
					t.Fatal("No request was made")
				}
				expectedMethod := "POST"
				if tt.cfg.UserConfig.PiphosGistID != "" {
					expectedMethod = "PATCH"
				}
				VerifyRequest(t, req, expectedMethod, url)
				body, _ := io.ReadAll(req.Body)
				var gist GithubGist
				if err := json.Unmarshal(body, &gist); err != nil {
					t.Fatalf("Failed to parse request body: %v", err)
				}
				if gist.Files[tt.cfg.UserConfig.Hostname].Content != tt.ip {
					t.Errorf("pushTender() ip in payload = %v, want %v",
						gist.Files[tt.cfg.UserConfig.Hostname].Content, tt.ip)
				}
			}
		})
	}
}

func TestPullTender(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		tender  Tender
		status  int
		body    string
		wantErr bool
	}{
		{
			name: "successful pull",
			cfg: Config{
				UserConfig: UserConfig{
					PiphosGistID: TestGistID,
				},
			},
			tender: Tender{
				Name: TenderGithub,
				URL:  "https://api.github.com/gists",
				Headers: map[string]string{
					"Authorization": "Bearer " + ValidClassicPAT,
				},
			},
			status: 200,
			body: NewGistResponse("", map[string]GithubFile{
				PayloadDescription: {Filename: PayloadDescription, Content: PayloadDescription},
				TestHostname:       {Filename: TestHostname, Content: TestIP},
			}),
			wantErr: false,
		},
		{
			name: "no gist ID configured",
			cfg: Config{
				UserConfig: UserConfig{},
			},
			tender: Tender{
				Name: TenderGithub,
				URL:  "https://api.github.com/gists",
			},
			status:  200,
			body:    "",
			wantErr: true,
		},
		{
			name: "server error",
			cfg: Config{
				UserConfig: UserConfig{
					PiphosGistID: TestGistID,
				},
			},
			tender: Tender{
				Name: TenderGithub,
				URL:  "https://api.github.com/gists",
				Headers: map[string]string{
					"Authorization": "Bearer " + ValidClassicPAT,
				},
			},
			status:  404,
			body:    "Not Found",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := NewMockTransport()
			url := tt.tender.URL
			if tt.cfg.UserConfig.PiphosGistID != "" {
				url += "/" + tt.cfg.UserConfig.PiphosGistID
			}
			transport.SetResponse(tt.status, tt.body)
			tt.cfg.Client = &http.Client{Transport: transport}

			_, err := pullTender(tt.cfg, tt.tender)
			if (err != nil) != tt.wantErr {
				t.Errorf("pullTender() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				req := transport.LastRequest()
				if req == nil {
					t.Fatal("No request was made")
				}
				expectedURL := tt.tender.URL + "/" + tt.cfg.UserConfig.PiphosGistID
				VerifyRequest(t, req, "GET", expectedURL)
			}
		})
	}
}
