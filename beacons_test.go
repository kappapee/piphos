package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"
)

type mockTransport struct {
	response *http.Response
	err      error
	requests []*http.Request
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	m.requests = append(m.requests, req)
	return m.response, m.err
}

func TestContactBeacon(t *testing.T) {
	originalBeaconConfig := BeaconConfig
	defer func() {
		BeaconConfig = originalBeaconConfig
	}()

	tests := []struct {
		name          string
		beaconID      string
		responseCode  int
		responseBody  string
		beaconConfigs map[string]Beacon
		wantIP        string
		wantErr       bool
	}{
		{
			name:         "Valid IP from AWS beacon",
			beaconID:     BeaconAws,
			responseCode: http.StatusOK,
			responseBody: "203.0.113.1\n",
			beaconConfigs: map[string]Beacon{
				BeaconAws: {Name: "aws", URL: "https://checkip.amazonaws.com"},
			},
			wantIP:  "203.0.113.1",
			wantErr: false,
		},
		{
			name:         "Valid IP from HAZ beacon",
			beaconID:     BeaconHaz,
			responseCode: http.StatusOK,
			responseBody: "203.0.113.2\n",
			beaconConfigs: map[string]Beacon{
				BeaconHaz: {Name: "icanhazip", URL: "https://ipv4.icanhazip.com"},
			},
			wantIP:  "203.0.113.2",
			wantErr: false,
		},
		{
			name:         "Invalid IP response",
			beaconID:     BeaconAws,
			responseCode: http.StatusOK,
			responseBody: "invalid-ip\n",
			beaconConfigs: map[string]Beacon{
				BeaconAws: {Name: "aws", URL: "https://checkip.amazonaws.com"},
			},
			wantIP:  "",
			wantErr: true,
		},
		{
			name:         "Server error response",
			beaconID:     BeaconAws,
			responseCode: http.StatusInternalServerError,
			responseBody: "Internal Server Error",
			beaconConfigs: map[string]Beacon{
				BeaconAws: {Name: "aws", URL: "https://checkip.amazonaws.com"},
			},
			wantIP:  "",
			wantErr: true,
		},
		{
			name:          "No beacons configured",
			beaconID:      BeaconAws,
			responseCode:  http.StatusOK,
			responseBody:  "203.0.113.1\n",
			beaconConfigs: map[string]Beacon{},
			wantIP:        "",
			wantErr:       true,
		},
		{
			name:         "Random beacon selection",
			beaconID:     BeaconDefault,
			responseCode: http.StatusOK,
			responseBody: "203.0.113.3\n",
			beaconConfigs: map[string]Beacon{
				BeaconAws: {Name: "aws", URL: "https://checkip.amazonaws.com"},
				BeaconHaz: {Name: "icanhazip", URL: "https://ipv4.icanhazip.com"},
			},
			wantIP:  "203.0.113.3",
			wantErr: false,
		},
		{
			name:         "Empty response",
			beaconID:     BeaconAws,
			responseCode: http.StatusOK,
			responseBody: "",
			beaconConfigs: map[string]Beacon{
				BeaconAws: {Name: "aws", URL: "https://checkip.amazonaws.com"},
			},
			wantIP:  "",
			wantErr: true,
		},
		{
			name:         "Response with whitespace",
			beaconID:     BeaconAws,
			responseCode: http.StatusOK,
			responseBody: "  203.0.113.4  \n\t",
			beaconConfigs: map[string]Beacon{
				BeaconAws: {Name: "aws", URL: "https://checkip.amazonaws.com"},
			},
			wantIP:  "203.0.113.4",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := &mockTransport{
				response: &http.Response{
					StatusCode: tt.responseCode,
					Body:       io.NopCloser(strings.NewReader(tt.responseBody)),
				},
			}

			cfg := Config{
				Client: &http.Client{
					Transport: transport,
					Timeout:   10 * time.Second,
				},
			}

			BeaconConfig = tt.beaconConfigs

			gotIP, err := contactBeacon(cfg, tt.beaconID)

			if (err != nil) != tt.wantErr {
				t.Errorf("contactBeacon() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && gotIP != tt.wantIP {
				t.Errorf("contactBeacon() = %v, want %v", gotIP, tt.wantIP)
			}

			if !tt.wantErr && len(transport.requests) > 0 {
				req := transport.requests[0]
				if !strings.HasPrefix(req.Header.Get("User-Agent"), "piphos/") {
					t.Error("User-Agent header should start with 'piphos/'")
				}
			}
		})
	}
}

func TestNetworkErrors(t *testing.T) {
	originalBeaconConfig := BeaconConfig
	defer func() {
		BeaconConfig = originalBeaconConfig
	}()

	BeaconConfig = map[string]Beacon{
		BeaconAws: {Name: "aws", URL: "https://test-aws.example.com"},
	}

	tests := []struct {
		name    string
		err     error
		wantErr string
	}{
		{
			name:    "Timeout error",
			err:     fmt.Errorf("timeout"),
			wantErr: "unable to get response",
		},
		{
			name:    "Connection error",
			err:     fmt.Errorf("connection refused"),
			wantErr: "unable to get response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := &mockTransport{
				err: tt.err,
			}

			cfg := Config{
				Client: &http.Client{
					Transport: transport,
					Timeout:   10 * time.Second,
				},
			}

			_, err := contactBeacon(cfg, BeaconAws)
			if err == nil {
				t.Fatal("contactBeacon() expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("contactBeacon() error = %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestLogOutput(t *testing.T) {
	originalBeaconConfig := BeaconConfig
	defer func() {
		BeaconConfig = originalBeaconConfig
	}()

	BeaconConfig = map[string]Beacon{
		"test1": {Name: "test1", URL: "https://test1.example.com"},
		"test2": {Name: "test2", URL: "https://test2.example.com"},
	}

	var logBuf strings.Builder
	log.SetOutput(&logBuf)
	defer log.SetOutput(io.Discard)

	transport := &mockTransport{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("203.0.113.1\n")),
		},
	}

	cfg := Config{
		Client: &http.Client{
			Transport: transport,
			Timeout:   10 * time.Second,
		},
	}

	_, err := contactBeacon(cfg, "")
	if err != nil {
		t.Fatalf("contactBeacon() error = %v", err)
	}

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "selecting random beacon") {
		t.Error("Expected log message about random beacon selection")
	}
}
