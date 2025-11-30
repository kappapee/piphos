package tender

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kappapee/piphos/internal/config"
)

func TestGithubPull_NoGist(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify authentication header
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			t.Error("Authorization header not set correctly")
		}

		// Return empty gist list
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	gh := newGithub("test-token")
	gh.baseURL = server.URL

	ctx := context.Background()
	result, err := gh.Pull(ctx)

	if err != nil {
		t.Errorf("expected no error but got: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result but got: %v", result)
	}
}

func TestGithubPull_WithGist(t *testing.T) {
	gistID := "test-gist-id"
	hostIPMap := map[string]string{
		"host1": "203.0.113.1",
		"host2": "203.0.113.2",
	}
	content, _ := json.Marshal(hostIPMap)

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		if callCount == 1 {
			// First call: list gists
			gists := []gist{
				{
					ID:          gistID,
					Description: config.PiphosStamp,
				},
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(gists)
		} else if callCount == 2 {
			// Second call: get specific gist
			gistResponse := gist{
				ID:          gistID,
				Description: config.PiphosStamp,
				Files: map[string]gistFile{
					config.PiphosStamp: {
						Content:   string(content),
						Filename:  config.PiphosStamp,
						Truncated: false,
					},
				},
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(gistResponse)
		}
	}))
	defer server.Close()

	gh := newGithub("test-token")
	gh.baseURL = server.URL

	ctx := context.Background()
	result, err := gh.Pull(ctx)

	if err != nil {
		t.Errorf("expected no error but got: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result but got nil")
	}
	if len(result) != 2 {
		t.Errorf("expected 2 entries but got %d", len(result))
	}
	if result["host1"] != "203.0.113.1" {
		t.Errorf("expected host1 IP to be 203.0.113.1 but got %s", result["host1"])
	}
	if result["host2"] != "203.0.113.2" {
		t.Errorf("expected host2 IP to be 203.0.113.2 but got %s", result["host2"])
	}
}

func TestGithubPull_TruncatedGist(t *testing.T) {
	gistID := "test-gist-id"

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		if callCount == 1 {
			// First call: list gists
			gists := []gist{
				{
					ID:          gistID,
					Description: config.PiphosStamp,
				},
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(gists)
		} else if callCount == 2 {
			// Second call: get specific gist with truncated file
			gistResponse := gist{
				ID:          gistID,
				Description: config.PiphosStamp,
				Files: map[string]gistFile{
					config.PiphosStamp: {
						Content:   `{"host1": "203.0.113.1"}`,
						Filename:  config.PiphosStamp,
						Truncated: true,
					},
				},
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(gistResponse)
		}
	}))
	defer server.Close()

	gh := newGithub("test-token")
	gh.baseURL = server.URL

	ctx := context.Background()
	_, err := gh.Pull(ctx)

	if err == nil {
		t.Error("expected error for truncated gist but got nil")
	}
	if !strings.Contains(err.Error(), "truncated") {
		t.Errorf("expected error message to mention truncation but got: %v", err)
	}
}

func TestGithubPull_MissingFile(t *testing.T) {
	gistID := "test-gist-id"

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		if callCount == 1 {
			// First call: list gists
			gists := []gist{
				{
					ID:          gistID,
					Description: config.PiphosStamp,
				},
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(gists)
		} else if callCount == 2 {
			// Second call: get specific gist without the expected file
			gistResponse := gist{
				ID:          gistID,
				Description: config.PiphosStamp,
				Files:       map[string]gistFile{},
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(gistResponse)
		}
	}))
	defer server.Close()

	gh := newGithub("test-token")
	gh.baseURL = server.URL

	ctx := context.Background()
	_, err := gh.Pull(ctx)

	if err == nil {
		t.Error("expected error for missing file but got nil")
	}
	if !strings.Contains(err.Error(), "missing file") {
		t.Errorf("expected error message to mention missing file but got: %v", err)
	}
}

func TestGithubPush_CreateNewGist(t *testing.T) {
	hostname := "testhost"
	ip := "203.0.113.1"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// No existing gist
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("[]"))
		} else if r.Method == http.MethodPost {
			// Verify the request body
			var payload gist
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("failed to decode request body: %v", err)
			}

			if payload.Description != config.PiphosStamp {
				t.Errorf("expected description %s but got %s", config.PiphosStamp, payload.Description)
			}
			if payload.Public {
				t.Error("expected private gist but got public")
			}

			file, ok := payload.Files[config.PiphosStamp]
			if !ok {
				t.Error("expected file not found in payload")
			}

			var content map[string]string
			if err := json.Unmarshal([]byte(file.Content), &content); err != nil {
				t.Errorf("failed to decode file content: %v", err)
			}

			if content[hostname] != ip {
				t.Errorf("expected IP %s for host %s but got %s", ip, hostname, content[hostname])
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(payload)
		}
	}))
	defer server.Close()

	gh := newGithub("test-token")
	gh.baseURL = server.URL

	ctx := context.Background()
	err := gh.Push(ctx, hostname, ip)

	if err != nil {
		t.Errorf("expected no error but got: %v", err)
	}
}

func TestGithubPush_UpdateExistingGist(t *testing.T) {
	gistID := "test-gist-id"
	hostname := "testhost"
	newIP := "203.0.113.2"
	existingContent := map[string]string{
		"otherhost": "203.0.113.1",
	}
	content, _ := json.Marshal(existingContent)

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		if r.Method == http.MethodGet && callCount == 1 {
			// First call: list gists
			gists := []gist{
				{
					ID:          gistID,
					Description: config.PiphosStamp,
				},
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(gists)
		} else if r.Method == http.MethodGet && callCount == 2 {
			// Second call: get specific gist
			gistResponse := gist{
				ID:          gistID,
				Description: config.PiphosStamp,
				Files: map[string]gistFile{
					config.PiphosStamp: {
						Content:   string(content),
						Filename:  config.PiphosStamp,
						Truncated: false,
					},
				},
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(gistResponse)
		} else if r.Method == http.MethodPatch {
			// Update gist
			var payload gist
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("failed to decode request body: %v", err)
			}

			file := payload.Files[config.PiphosStamp]
			var updatedContent map[string]string
			if err := json.Unmarshal([]byte(file.Content), &updatedContent); err != nil {
				t.Errorf("failed to decode file content: %v", err)
			}

			if updatedContent[hostname] != newIP {
				t.Errorf("expected new IP %s for host %s but got %s", newIP, hostname, updatedContent[hostname])
			}
			if updatedContent["otherhost"] != "203.0.113.1" {
				t.Error("expected existing host to remain unchanged")
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(payload)
		}
	}))
	defer server.Close()

	gh := newGithub("test-token")
	gh.baseURL = server.URL

	ctx := context.Background()
	err := gh.Push(ctx, hostname, newIP)

	if err != nil {
		t.Errorf("expected no error but got: %v", err)
	}
}

func TestGithubPush_SkipUnchangedIP(t *testing.T) {
	gistID := "test-gist-id"
	hostname := "testhost"
	sameIP := "203.0.113.1"
	existingContent := map[string]string{
		hostname: sameIP,
	}
	content, _ := json.Marshal(existingContent)

	callCount := 0
	patchCalled := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		if r.Method == http.MethodGet && callCount == 1 {
			// First call: list gists
			gists := []gist{
				{
					ID:          gistID,
					Description: config.PiphosStamp,
				},
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(gists)
		} else if r.Method == http.MethodGet && callCount == 2 {
			// Second call: get specific gist
			gistResponse := gist{
				ID:          gistID,
				Description: config.PiphosStamp,
				Files: map[string]gistFile{
					config.PiphosStamp: {
						Content:   string(content),
						Filename:  config.PiphosStamp,
						Truncated: false,
					},
				},
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(gistResponse)
		} else if r.Method == http.MethodPatch {
			patchCalled = true
			t.Error("PATCH should not be called when IP hasn't changed")
		}
	}))
	defer server.Close()

	gh := newGithub("test-token")
	gh.baseURL = server.URL

	ctx := context.Background()
	err := gh.Push(ctx, hostname, sameIP)

	if err != nil {
		t.Errorf("expected no error but got: %v", err)
	}
	if patchCalled {
		t.Error("expected no PATCH request when IP unchanged")
	}
}

func TestGithubPush_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message": "Bad credentials"}`))
	}))
	defer server.Close()

	gh := newGithub("invalid-token")
	gh.baseURL = server.URL

	ctx := context.Background()
	err := gh.Push(ctx, "testhost", "203.0.113.1")

	if err == nil {
		t.Error("expected error but got nil")
	}
	if !strings.Contains(err.Error(), "unexpected response status") {
		t.Errorf("expected status error but got: %v", err)
	}
}

func TestGithubRequestTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	gh := newGithub("test-token")
	gh.baseURL = server.URL

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := gh.Pull(ctx)
	if err == nil {
		t.Error("expected timeout error but got nil")
	}
}

func TestGithubInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	gh := newGithub("test-token")
	gh.baseURL = server.URL

	ctx := context.Background()
	_, err := gh.Pull(ctx)

	if err == nil {
		t.Error("expected JSON unmarshal error but got nil")
	}
}

func TestGithubPull_InvalidFileContent(t *testing.T) {
	gistID := "test-gist-id"

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		if callCount == 1 {
			// First call: list gists
			gists := []gist{
				{
					ID:          gistID,
					Description: config.PiphosStamp,
				},
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(gists)
		} else if callCount == 2 {
			// Second call: get specific gist with invalid JSON content
			gistResponse := gist{
				ID:          gistID,
				Description: config.PiphosStamp,
				Files: map[string]gistFile{
					config.PiphosStamp: {
						Content:   "invalid json content",
						Filename:  config.PiphosStamp,
						Truncated: false,
					},
				},
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(gistResponse)
		}
	}))
	defer server.Close()

	gh := newGithub("test-token")
	gh.baseURL = server.URL

	ctx := context.Background()
	_, err := gh.Pull(ctx)

	if err == nil {
		t.Error("expected error for invalid file content but got nil")
	}
	if !strings.Contains(err.Error(), "unmarshal") {
		t.Errorf("expected unmarshal error but got: %v", err)
	}
}

func TestGithubCreateGist_Success(t *testing.T) {
	// This test verifies createGist works correctly

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// No existing gist
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("[]"))
		} else if r.Method == http.MethodPost {
			// Create gist
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id": "new-gist-id"}`))
		}
	}))
	defer server.Close()

	gh := newGithub("test-token")
	gh.baseURL = server.URL

	ctx := context.Background()

	// Push will call createGist when no gist exists
	err := gh.Push(ctx, "testhost", "203.0.113.1")

	// Should succeed
	if err != nil {
		t.Errorf("expected no error but got: %v", err)
	}
}

func TestGithubHeaders(t *testing.T) {
	expectedHeaders := map[string]string{
		"User-Agent":           config.PiphosUserAgent,
		"Accept":               "application/vnd.github+json",
		"X-GitHub-Api-Version": "2022-11-28",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify all required headers are set
		for key, expectedValue := range expectedHeaders {
			if r.Header.Get(key) != expectedValue {
				t.Errorf("expected header %s to be %s but got %s", key, expectedValue, r.Header.Get(key))
			}
		}

		// Verify Authorization header format
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer test-token") {
			t.Errorf("expected Authorization header to start with 'Bearer test-token' but got %s", auth)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	gh := newGithub("test-token")
	gh.baseURL = server.URL

	ctx := context.Background()
	gh.Pull(ctx)
}
