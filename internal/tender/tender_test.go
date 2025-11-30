package tender

import (
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name          string
		tender        string
		token         string
		expectedError bool
	}{
		{
			name:          "gh tender with valid token",
			tender:        "gh",
			token:         "valid-token",
			expectedError: false,
		},
		{
			name:          "gh tender with empty token",
			tender:        "gh",
			token:         "",
			expectedError: true,
		},
		{
			name:          "unknown tender",
			tender:        "unknown",
			token:         "valid-token",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable for the test
			if tt.token != "" {
				os.Setenv("PIPHOS_GITHUB_TOKEN", tt.token)
			} else {
				os.Unsetenv("PIPHOS_GITHUB_TOKEN")
			}
			defer os.Unsetenv("PIPHOS_GITHUB_TOKEN")

			tender, err := New(tt.tender)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error but got nil")
				}
				if tender != nil {
					t.Error("expected nil tender but got non-nil")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
				if tender == nil {
					t.Error("expected non-nil tender but got nil")
				}
			}
		})
	}
}
