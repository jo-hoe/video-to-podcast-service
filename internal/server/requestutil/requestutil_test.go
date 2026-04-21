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
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			result := BaseURL(ctx)

			assert.Equal(t, tt.wantScheme, result.Scheme)
			assert.Equal(t, tt.wantHost, result.Host)
		})
	}
}
