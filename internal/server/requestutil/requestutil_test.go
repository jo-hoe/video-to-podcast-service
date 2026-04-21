package requestutil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestBaseURL(t *testing.T) {
	tests := []struct {
		name           string
		host           string
		forwardedProto string
		forwardedHost  string
		forwarded      string
		wantScheme     string
		wantHost       string
	}{
		{
			name:       "direct access without proxy headers",
			host:       "192.168.1.100:8080",
			wantScheme: "http",
			wantHost:   "192.168.1.100:8080",
		},
		{
			name:           "https via X-Forwarded-Proto",
			host:           "internal-service:8080",
			forwardedProto: "https",
			wantScheme:     "https",
			wantHost:       "internal-service:8080",
		},
		{
			name:          "external host via X-Forwarded-Host",
			host:          "internal-service:8080",
			forwardedHost: "podcast.example.com",
			wantScheme:    "http",
			wantHost:      "podcast.example.com",
		},
		{
			name:           "both forwarded proto and host",
			host:           "internal-service:8080",
			forwardedProto: "https",
			forwardedHost:  "podcast.example.com",
			wantScheme:     "https",
			wantHost:       "podcast.example.com",
		},
		{
			name:       "RFC 7239 Forwarded header with proto and host",
			host:       "internal-service:8080",
			forwarded:  "for=192.0.2.60;proto=https;host=podcast.example.com",
			wantScheme: "https",
			wantHost:   "podcast.example.com",
		},
		{
			name:       "RFC 7239 Forwarded header with proto only",
			host:       "internal-service:8080",
			forwarded:  "for=192.0.2.60;proto=https",
			wantScheme: "https",
			wantHost:   "internal-service:8080",
		},
		{
			name:       "RFC 7239 Forwarded header with host only",
			host:       "internal-service:8080",
			forwarded:  "for=192.0.2.60;host=podcast.example.com",
			wantScheme: "http",
			wantHost:   "podcast.example.com",
		},
		{
			name:           "RFC 7239 takes precedence over X-Forwarded headers",
			host:           "internal-service:8080",
			forwardedProto: "http",
			forwardedHost:  "old.example.com",
			forwarded:      "proto=https;host=new.example.com",
			wantScheme:     "https",
			wantHost:       "new.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Host = tt.host
			if tt.forwardedProto != "" {
				req.Header.Set("X-Forwarded-Proto", tt.forwardedProto)
			}
			if tt.forwardedHost != "" {
				req.Header.Set("X-Forwarded-Host", tt.forwardedHost)
			}
			if tt.forwarded != "" {
				req.Header.Set("Forwarded", tt.forwarded)
			}
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			result := BaseURL(ctx)

			assert.Equal(t, tt.wantScheme, result.Scheme)
			assert.Equal(t, tt.wantHost, result.Host)
		})
	}
}

func TestParseForwarded(t *testing.T) {
	tests := []struct {
		name       string
		header     string
		wantScheme string
		wantHost   string
	}{
		{
			name:       "empty header",
			header:     "",
			wantScheme: "",
			wantHost:   "",
		},
		{
			name:       "proto and host",
			header:     "for=192.0.2.60;proto=https;host=example.com",
			wantScheme: "https",
			wantHost:   "example.com",
		},
		{
			name:       "proto only",
			header:     "proto=https",
			wantScheme: "https",
			wantHost:   "",
		},
		{
			name:       "host only",
			header:     "host=example.com",
			wantScheme: "",
			wantHost:   "example.com",
		},
		{
			name:       "multiple proxies uses first element",
			header:     "host=first.example.com;proto=https, host=second.example.com;proto=http",
			wantScheme: "https",
			wantHost:   "first.example.com",
		},
		{
			name:       "case insensitive keys",
			header:     "Proto=https;Host=example.com",
			wantScheme: "https",
			wantHost:   "example.com",
		},
		{
			name:       "whitespace around values",
			header:     " proto = https ; host = example.com ",
			wantScheme: "https",
			wantHost:   "example.com",
		},
		{
			name:       "unknown directives are ignored",
			header:     "for=192.0.2.60;by=proxy;proto=https;host=example.com",
			wantScheme: "https",
			wantHost:   "example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme, host := parseForwarded(tt.header)
			assert.Equal(t, tt.wantScheme, scheme)
			assert.Equal(t, tt.wantHost, host)
		})
	}
}
