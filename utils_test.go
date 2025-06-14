package main

import (
	"testing"
)

func TestValidateIP(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"127.0.0.1", false},
		{"::1", false},
		{"255.255.255.255", false},
		{"256.256.256.256", true},
		{"hello", true},
		{"", true},
		{"not.an.ip.add", true},
		{"::1", false},
		{"2001:0db8:85a3:0000:0000:8a2e:0370:7334", false},
		{"2001:db8:85a3::8a2e:370:7334", false},
		{"fe80::", false},
		{"::", false},
		{"2001:db8:85a3::8a2e:370g:7334", true},
		{"2001:db8:85a3::8a2e:370:7334::", true},
		{"12345::abcd", true},
		{"::1::", true},
	}

	for _, tt := range tests {
		err := validateIP(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("validateIP(%q) error = %v; wantErr = %v", tt.input, err, tt.wantErr)
		}
	}
}

func TestValidateToken(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		tender  string
		wantErr bool
	}{
		{
			name:    "Valid classic PAT",
			token:   "ghp_1234567890abcdef1234567890abcdef1234",
			tender:  "github",
			wantErr: false,
		},
		{
			name:    "Valid fine-grained PAT",
			token:   "github_pat_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_0123456789_abcdef",
			tender:  "github",
			wantErr: false,
		},
		{
			name:    "Empty token",
			token:   "",
			tender:  "github",
			wantErr: true,
		},
		{
			name:    "Invalid classic PAT (too short)",
			token:   "ghp_1234",
			tender:  "github",
			wantErr: true,
		},
		{
			name:    "Invalid fine-grained PAT (wrong prefix)",
			token:   "githubpat_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_0123456789_abcdef",
			tender:  "github",
			wantErr: true,
		},
		{
			name:    "Unknown tender name",
			token:   "ghp_1234567890abcdef1234567890abcdef1234",
			tender:  "gitlab",
			wantErr: true,
		},
		{
			name:    "Classic PAT with spaces",
			token:   "   ghp_1234567890abcdef1234567890abcdef1234   ",
			tender:  "github",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateToken(tt.token, tt.tender)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateToken(%q, %q) error = %v, wantErr %v", tt.token, tt.tender, err, tt.wantErr)
			}
		})
	}
}
