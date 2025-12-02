package exec

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestPing(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedError bool
	}{
		{
			name:          "unknown beacon",
			args:          []string{"-beacon", "unknown"},
			expectedError: true,
		},
		{
			name:          "extra arguments",
			args:          []string{"extra"},
			expectedError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := Ping(ctx, tt.args)
			if tt.expectedError {
				if err == nil {
					t.Error("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestPull(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedError bool
	}{
		{
			name:          "unknown tender",
			args:          []string{"-tender", "unknown"},
			expectedError: true,
		},
		{
			name:          "extra arguments",
			args:          []string{"extra"},
			expectedError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := Pull(ctx, tt.args)
			if tt.expectedError {
				if err == nil {
					t.Error("expected error but got nil")
				}
			}
		})
	}
}

func TestPullMissingToken(t *testing.T) {
	os.Unsetenv("PIPHOS_GITHUB_TOKEN")
	ctx := context.Background()
	_, err := Pull(ctx, []string{})
	if err == nil {
		t.Error("expected error for missing token but got nil")
	}
	if !strings.Contains(err.Error(), "token") {
		t.Errorf("expected token error but got: %v", err)
	}
}

func TestPush(t *testing.T) {
	os.Setenv("PIPHOS_GITHUB_TOKEN", "test-token")
	defer os.Unsetenv("PIPHOS_GITHUB_TOKEN")
	tests := []struct {
		name          string
		args          []string
		expectedError bool
	}{
		{
			name:          "unknown tender",
			args:          []string{"-tender", "unknown"},
			expectedError: true,
		},
		{
			name:          "unknown beacon",
			args:          []string{"-beacon", "unknown"},
			expectedError: true,
		},
		{
			name:          "extra arguments",
			args:          []string{"extra"},
			expectedError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := Push(ctx, tt.args)
			if tt.expectedError {
				if err == nil {
					t.Error("expected error but got nil")
				}
			}
		})
	}
}

func TestPushMissingToken(t *testing.T) {
	os.Unsetenv("PIPHOS_GITHUB_TOKEN")
	ctx := context.Background()
	err := Push(ctx, []string{})
	if err == nil {
		t.Error("expected error for missing token but got nil")
	}
	if !strings.Contains(err.Error(), "token") {
		t.Errorf("expected token error but got: %v", err)
	}
}
