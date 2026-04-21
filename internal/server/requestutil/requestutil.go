package requestutil

import (
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
)

func BaseURL(ctx echo.Context) *url.URL {
	scheme, host := parseForwarded(ctx.Request().Header.Get("Forwarded"))

	if scheme == "" {
		scheme = ctx.Scheme()
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
