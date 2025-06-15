package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
)

const (
	ValidClassicPAT     = "ghp_1234567890abcdef1234567890abcdef1234"
	ValidFineGrainedPAT = "github_pat_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_0123456789_abcdef"
	TestHostname        = "test-host"
	TestGistID          = "test-gist-123"
	TestIP              = "203.0.113.1"
)

type MockTransport struct {
	mu       sync.Mutex
	requests []*http.Request
	response *http.Response
	err      error
}

type MockResponse struct {
	Headers    map[string]string
	Body       string
	StatusCode int
	Error      error
}

func NewTestConfig(transport http.RoundTripper) Config {
	return Config{
		Client: &http.Client{Transport: transport},
		UserConfig: UserConfig{
			Hostname: TestHostname,
			Token:    ValidClassicPAT,
		},
	}
}

func NewMockTransport() *MockTransport {
	return &MockTransport{
		response: &http.Response{
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("{}")),
			StatusCode: http.StatusOK,
		},
	}
}

func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests = append(m.requests, req)
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

func (m *MockTransport) SetResponse(statusCode int, body string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.response = &http.Response{
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		StatusCode: statusCode,
	}
}

func (m *MockTransport) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.err = err
}

func (m *MockTransport) LastRequest() *http.Request {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.requests) == 0 {
		return nil
	}
	return m.requests[len(m.requests)-1]
}

func (m *MockTransport) ClearRequests() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests = nil
}

func JSONResponse(data any) string {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal test data: %v", err))
	}
	return string(b)
}

func MockResponseOK(body string) MockResponse {
	return MockResponse{
		Headers:    make(map[string]string),
		Body:       body,
		StatusCode: http.StatusOK,
	}
}

func MockResponseError(statusCode int, message string) MockResponse {
	return MockResponse{
		Headers:    make(map[string]string),
		Body:       message,
		StatusCode: statusCode,
	}
}

func VerifyRequest(t *testing.T, req *http.Request, method, url string) {
	t.Helper()
	if req == nil {
		t.Fatal("No request was made")
	}
	if req.Method != method {
		t.Errorf("Expected method %s, got %s", method, req.Method)
	}
	if req.URL.String() != url {
		t.Errorf("Expected URL %s, got %s", url, req.URL.String())
	}
}
