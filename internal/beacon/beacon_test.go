package beacon

import (
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name          string
		beacon        string
		expectedError bool
	}{
		{
			name:          "haz beacon",
			beacon:        "haz",
			expectedError: false,
		},
		{
			name:          "aws beacon",
			beacon:        "aws",
			expectedError: false,
		},
		{
			name:          "unknown beacon",
			beacon:        "unknown",
			expectedError: true,
		},
		{
			name:          "empty beacon",
			beacon:        "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := New(tt.beacon)
			if tt.expectedError {
				if err == nil {
					t.Error("expected error but got nil")
				}
				if b != nil {
					t.Error("expected nil beacon but got non-nil")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
				if b == nil {
					t.Error("expected non-nil beacon but got nil")
				}
			}
		})
	}
}

func TestNewBeaconURLs(t *testing.T) {
	tests := []struct {
		name         string
		beacon       string
		expectedURL  string
		expectedName string
	}{
		{
			name:         "haz beacon URL",
			beacon:       "haz",
			expectedURL:  "https://icanhazip.com",
			expectedName: "haz",
		},
		{
			name:         "aws beacon URL",
			beacon:       "aws",
			expectedURL:  "https://checkip.amazonaws.com",
			expectedName: "aws",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := New(tt.beacon)
			if err != nil {
				t.Fatalf("expected no error but got: %v", err)
			}

			// Type assert to access internal fields for verification
			if web, ok := b.(*web); ok {
				if web.baseURL != tt.expectedURL {
					t.Errorf("expected baseURL %s but got %s", tt.expectedURL, web.baseURL)
				}
				if web.name != tt.expectedName {
					t.Errorf("expected name %s but got %s", tt.expectedName, web.name)
				}
			} else {
				t.Error("expected beacon to be of type *web")
			}
		})
	}
}
