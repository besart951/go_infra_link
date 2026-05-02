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
		{name: "different host", host: "example.test", origin: "https://evil.test", want: false},
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

			if got := SameHostOrigin(req); got != tt.want {
				t.Fatalf("SameHostOrigin() = %v, want %v", got, tt.want)
			}
		})
	}
}
