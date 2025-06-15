package main

import (
	"fmt"
	"strings"
	"testing"
)

func NewBeaconResponse(ip string) string {
	return ip + "\n"
}

func TestContactBeacon(t *testing.T) {
	originalBeaconConfig := BeaconConfig
	defer func() { BeaconConfig = originalBeaconConfig }()

	tests := []struct {
		name          string
		beaconID      string
		statusCode    int
		body          string
		beaconConfigs map[string]Beacon
		wantIP        string
		wantErr       bool
	}{
		{
			name:       "Valid IP from AWS beacon",
			beaconID:   BeaconAws,
			statusCode: 200,
			body:       NewBeaconResponse("203.0.113.1"),
			beaconConfigs: map[string]Beacon{
				BeaconAws: {Name: "aws", URL: "https://checkip.amazonaws.com"},
			},
			wantIP:  "203.0.113.1",
			wantErr: false,
		},
		{
			name:       "Valid IP from HAZ beacon",
			beaconID:   BeaconHaz,
			statusCode: 200,
			body:       NewBeaconResponse("203.0.113.2"),
			beaconConfigs: map[string]Beacon{
				BeaconHaz: {Name: "icanhazip", URL: "https://ipv4.icanhazip.com"},
			},
			wantIP:  "203.0.113.2",
			wantErr: false,
		},
		{
			name:       "Invalid IP response",
			beaconID:   BeaconAws,
			statusCode: 200,
			body:       "invalid-ip\n",
			beaconConfigs: map[string]Beacon{
				BeaconAws: {Name: "aws", URL: "https://checkip.amazonaws.com"},
			},
			wantIP:  "",
			wantErr: true,
		},
		{
			name:       "Server error response",
			beaconID:   BeaconAws,
			statusCode: 500,
			body:       "Internal Server Error",
			beaconConfigs: map[string]Beacon{
				BeaconAws: {Name: "aws", URL: "https://checkip.amazonaws.com"},
			},
			wantIP:  "",
			wantErr: true,
		},
		{
			name:          "No beacons configured",
			beaconID:      BeaconAws,
			statusCode:    200,
			body:          NewBeaconResponse("203.0.113.1"),
			beaconConfigs: map[string]Beacon{},
			wantIP:        "",
			wantErr:       true,
		},
		{
			name:       "Random beacon selection",
			beaconID:   BeaconDefault,
			statusCode: 200,
			body:       NewBeaconResponse("203.0.113.3"),
			beaconConfigs: map[string]Beacon{
				BeaconAws: {Name: "aws", URL: "https://checkip.amazonaws.com"},
				BeaconHaz: {Name: "icanhazip", URL: "https://ipv4.icanhazip.com"},
			},
			wantIP:  "203.0.113.3",
			wantErr: false,
		},
		{
			name:       "Empty response",
			beaconID:   BeaconAws,
			statusCode: 200,
			body:       "",
			beaconConfigs: map[string]Beacon{
				BeaconAws: {Name: "aws", URL: "https://checkip.amazonaws.com"},
			},
			wantIP:  "",
			wantErr: true,
		},
		{
			name:       "Response with whitespace",
			beaconID:   BeaconAws,
			statusCode: 200,
			body:       "  203.0.113.4  \n\t",
			beaconConfigs: map[string]Beacon{
				BeaconAws: {Name: "aws", URL: "https://checkip.amazonaws.com"},
			},
			wantIP:  "203.0.113.4",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := NewMockTransport()
			transport.SetResponse(tt.statusCode, tt.body)
			cfg := NewTestConfig(transport)
			BeaconConfig = tt.beaconConfigs

			gotIP, err := contactBeacon(cfg, tt.beaconID)
			if (err != nil) != tt.wantErr {
				t.Errorf("contactBeacon() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && gotIP != tt.wantIP {
				t.Errorf("contactBeacon() = %v, want %v", gotIP, tt.wantIP)
			}
			req := transport.LastRequest()
			if req != nil && !strings.HasPrefix(req.Header.Get("User-Agent"), "piphos/") {
				t.Error("User-Agent header should start with 'piphos/'")
			}
		})
	}
}

func TestBeaconNetworkErrors(t *testing.T) {
	originalBeaconConfig := BeaconConfig
	defer func() { BeaconConfig = originalBeaconConfig }()

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
			transport := NewMockTransport()
			transport.SetError(tt.err)
			cfg := NewTestConfig(transport)

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
