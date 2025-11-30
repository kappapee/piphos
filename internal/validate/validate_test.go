package validate

import (
	"testing"
)

func TestCommand(t *testing.T) {
	tests := []struct {
		name          string
		nonFlagArgs   int
		expectedError bool
	}{
		{
			name:          "no extra arguments",
			nonFlagArgs:   0,
			expectedError: false,
		},
		{
			name:          "one extra argument",
			nonFlagArgs:   1,
			expectedError: true,
		},
		{
			name:          "multiple extra arguments",
			nonFlagArgs:   5,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Command(tt.nonFlagArgs)
			if tt.expectedError && err == nil {
				t.Errorf("expected error but got nil")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}

func TestIP(t *testing.T) {
	tests := []struct {
		name          string
		ip            string
		expectedError bool
	}{
		{
			name:          "valid IPv4",
			ip:            "192.168.1.1",
			expectedError: false,
		},
		{
			name:          "valid IPv4 public",
			ip:            "8.8.8.8",
			expectedError: false,
		},
		{
			name:          "valid IPv6",
			ip:            "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			expectedError: false,
		},
		{
			name:          "valid IPv6 compressed",
			ip:            "2001:db8::1",
			expectedError: false,
		},
		{
			name:          "valid IPv6 loopback",
			ip:            "::1",
			expectedError: false,
		},
		{
			name:          "invalid IP - empty string",
			ip:            "",
			expectedError: true,
		},
		{
			name:          "invalid IP - malformed",
			ip:            "256.256.256.256",
			expectedError: true,
		},
		{
			name:          "invalid IP - text",
			ip:            "not-an-ip",
			expectedError: true,
		},
		{
			name:          "invalid IP - partial",
			ip:            "192.168.1",
			expectedError: true,
		},
		{
			name:          "invalid IP - extra octet",
			ip:            "192.168.1.1.1",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IP(tt.ip)
			if tt.expectedError && err == nil {
				t.Errorf("expected error but got nil for IP: %s", tt.ip)
			}
			if !tt.expectedError && err != nil {
				t.Errorf("expected no error but got: %v for IP: %s", err, tt.ip)
			}
		})
	}
}

func TestToken(t *testing.T) {
	tests := []struct {
		name          string
		token         string
		expectedError bool
	}{
		{
			name:          "valid token",
			token:         "ghp_1234567890abcdef",
			expectedError: false,
		},
		{
			name:          "valid simple token",
			token:         "token",
			expectedError: false,
		},
		{
			name:          "empty token",
			token:         "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Token(tt.token)
			if tt.expectedError && err == nil {
				t.Errorf("expected error but got nil")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}
