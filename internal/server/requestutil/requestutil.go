package requestutil

import (
	"net/url"

	"github.com/labstack/echo/v4"
)

func BaseURL(ctx echo.Context) *url.URL {
	host := ctx.Request().Host
	if fwdHost := ctx.Request().Header.Get("X-Forwarded-Host"); fwdHost != "" {
		host = fwdHost
	}
	return &url.URL{Scheme: ctx.Scheme(), Host: host}
}
