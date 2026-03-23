package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name           string
		remoteAddr     string
		xForwardedFor  string
		xRealIP        string
		trustProxyHops int
		expected       string
	}{
		{
			name:           "direct connection",
			remoteAddr:     "192.168.1.1:12345",
			trustProxyHops: 0,
			expected:       "192.168.1.1",
		},
		{
			name:           "x-forwarded-for with trust",
			remoteAddr:     "10.0.0.1:12345",
			xForwardedFor:  "203.0.113.1, 198.51.100.1",
			trustProxyHops: 1,
			expected:       "198.51.100.1",
		},
		{
			name:           "x-real-ip",
			remoteAddr:     "10.0.0.1:12345",
			xRealIP:        "203.0.113.1",
			trustProxyHops: 1,
			expected:       "203.0.113.1",
		},
		{
			name:           "no trust proxy",
			remoteAddr:     "192.168.1.1:12345",
			xForwardedFor:  "203.0.113.1",
			trustProxyHops: 0,
			expected:       "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.RemoteAddr = tt.remoteAddr
			if tt.xForwardedFor != "" {
				r.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				r.Header.Set("X-Real-IP", tt.xRealIP)
			}

			result := GetClientIP(r, tt.trustProxyHops)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}
