package realtime

import (
	"net/http"
	"testing"
)

func TestSameHostOrigin(t *testing.T) {
	tests := []struct {
		name   string
		host   string
		origin string
		want   bool
	}{
		{name: "missing origin", host: "example.test", want: true},
		{name: "same host", host: "example.test", origin: "https://example.test", want: true},
		{name: "same host with port", host: "example.test:8080", origin: "https://example.test:8080", want: true},
		{name: "loopback aliases for dev proxy", host: "127.0.0.1:8080", origin: "http://localhost:5173", want: true},
		{name: "ipv6 loopback alias for dev proxy", host: "[::1]:8080", origin: "http://localhost:5173", want: true},
		{name: "forwarded host from reverse proxy", host: "backend:8080", origin: "https://app.example.test", want: true},
		{name: "different host", host: "example.test", origin: "https://evil.test", want: false},
		{name: "loopback origin does not match non-loopback host", host: "example.test", origin: "http://localhost:5173", want: false},
		{name: "invalid origin", host: "example.test", origin: "://bad", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "http://"+tt.host+"/ws", nil)
			if err != nil {
				t.Fatalf("request: %v", err)
			}
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if tt.name == "forwarded host from reverse proxy" {
				req.Header.Set("X-Forwarded-Host", "app.example.test")
			}

			if got := SameHostOrigin(req); got != tt.want {
				t.Fatalf("SameHostOrigin() = %v, want %v", got, tt.want)
			}
		})
	}
}
