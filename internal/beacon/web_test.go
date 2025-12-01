package beacon

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWebPing(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		responseStatus int
		expectedIP     string
		expectedError  bool
	}{
		{
			name:           "valid IPv4",
			responseBody:   "203.0.113.1",
			responseStatus: http.StatusOK,
			expectedIP:     "203.0.113.1",
			expectedError:  false,
		},
		{
			name:           "valid IPv4 with whitespace",
			responseBody:   "  203.0.113.1\n",
			responseStatus: http.StatusOK,
			expectedIP:     "203.0.113.1",
			expectedError:  false,
		},
		{
			name:           "valid IPv6",
			responseBody:   "2001:db8::1",
			responseStatus: http.StatusOK,
			expectedIP:     "2001:db8::1",
			expectedError:  false,
		},
		{
			name:           "invalid IP response",
			responseBody:   "not-an-ip",
			responseStatus: http.StatusOK,
			expectedError:  true,
		},
		{
			name:           "empty response",
			responseBody:   "",
			responseStatus: http.StatusOK,
			expectedError:  true,
		},
		{
			name:           "non-200 status",
			responseBody:   "203.0.113.1",
			responseStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
		{
			name:           "404 status",
			responseBody:   "not found",
			responseStatus: http.StatusNotFound,
			expectedError:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("User-Agent") == "" {
					t.Error("User-Agent header not set")
				}
				w.WriteHeader(tt.responseStatus)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()
			b := newWeb(server.URL, "test")
			ctx := context.Background()
			ip, err := b.Ping(ctx)
			if tt.expectedError {
				if err == nil {
					t.Error("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
				if ip != tt.expectedIP {
					t.Errorf("expected IP %s but got %s", tt.expectedIP, ip)
				}
			}
		})
	}
}

func TestWebPingTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("203.0.113.1"))
	}))
	defer server.Close()
	b := newWeb(server.URL, "test")
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := b.Ping(ctx)
	if err == nil {
		t.Error("expected timeout error but got nil")
	}
}

func TestWebPingCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("203.0.113.1"))
	}))
	defer server.Close()
	b := newWeb(server.URL, "test")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := b.Ping(ctx)
	if err == nil {
		t.Error("expected context cancellation error but got nil")
	}
}

func TestNewWeb(t *testing.T) {
	tests := []struct {
		name       string
		baseURL    string
		beaconName string
	}{
		{
			name:       "aws beacon",
			baseURL:    "https://checkip.amazonaws.com",
			beaconName: "aws",
		},
		{
			name:       "haz beacon",
			baseURL:    "https://icanhazip.com",
			beaconName: "haz",
		},
		{
			name:       "custom beacon",
			baseURL:    "https://example.com/ip",
			beaconName: "custom",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWeb(tt.baseURL, tt.beaconName)
			if w == nil {
				t.Fatal("expected non-nil web beacon")
			}
			if w.baseURL != tt.baseURL {
				t.Errorf("expected baseURL %s but got %s", tt.baseURL, w.baseURL)
			}
			if w.name != tt.beaconName {
				t.Errorf("expected name %s but got %s", tt.beaconName, w.name)
			}
			if w.client == nil {
				t.Error("expected non-nil HTTP client")
			}
			if w.headers == nil {
				t.Error("expected non-nil headers map")
			}
			if _, ok := w.headers["User-Agent"]; !ok {
				t.Error("expected User-Agent header to be set")
			}
		})
	}
}
