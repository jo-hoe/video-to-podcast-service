package requestutil

import (
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
)

// publicBaseURL, when non-nil, is used verbatim for all generated links and
// forwarded/host headers from the request are ignored. It is set once at
// startup via SetPublicBaseURL before any request is served.
var publicBaseURL *url.URL

// SetPublicBaseURL configures an explicit, trusted base URL for generated
// links. Passing an empty string clears it (falling back to request-derived
// values). It must be called during startup, before the server accepts
// requests. Returns an error if the provided value cannot be parsed.
func SetPublicBaseURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		publicBaseURL = nil
		return nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	publicBaseURL = &url.URL{Scheme: u.Scheme, Host: u.Host}
	return nil
}

func BaseURL(ctx echo.Context) *url.URL {
	// When an explicit public base URL is configured, trust it exclusively and
	// ignore client-supplied Forwarded / X-Forwarded-Host headers to prevent
	// feed-link/host poisoning.
	if publicBaseURL != nil {
		result := *publicBaseURL
		return &result
	}

	scheme, host := parseForwarded(ctx.Request().Header.Get("Forwarded"))

	if scheme == "" {
		if fwdProto := ctx.Request().Header.Get("X-Forwarded-Proto"); fwdProto != "" {
			scheme = fwdProto
		} else {
			scheme = ctx.Scheme()
		}
	}
	if host == "" {
		host = ctx.Request().Host
		if fwdHost := ctx.Request().Header.Get("X-Forwarded-Host"); fwdHost != "" {
			host = fwdHost
		}
	}

	return &url.URL{Scheme: scheme, Host: host}
}

func parseForwarded(header string) (scheme, host string) {
	if header == "" {
		return "", ""
	}
	// RFC 7239: Forwarded: for=...;host=example.com;proto=https, for=...
	// Use only the first element (closest proxy).
	if idx := strings.IndexByte(header, ','); idx != -1 {
		header = header[:idx]
	}
	for part := range strings.SplitSeq(header, ";") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(kv[0]))
		val := strings.TrimSpace(kv[1])
		switch key {
		case "proto":
			scheme = val
		case "host":
			host = val
		}
	}
	return scheme, host
}
