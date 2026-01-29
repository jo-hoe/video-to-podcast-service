package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/jo-hoe/video-to-podcast-service/internal/config"
	"github.com/jo-hoe/video-to-podcast-service/internal/core"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"
	"github.com/jo-hoe/video-to-podcast-service/internal/server/api"
	"github.com/jo-hoe/video-to-podcast-service/internal/server/ui"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var defaultResourcePath string

func StartServer(databaseService database.DatabaseService, cfg *config.Config) {
	defaultResourcePath = cfg.Persistence.Media.MediaPath

	e := echo.New()
	// Use RequestLogger with LogValuesFunc to satisfy linter and avoid panic.
	// Skip logging for probe and health endpoints.
	// Minimal custom request logger that reliably logs method, uri, status, latency, and user agent.
	// Skip logging for /health and /probe.
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			if path == "/health" || path == "/probe" {
				return next(c)
			}
			start := time.Now()
			err := next(c)
			req := c.Request()
			res := c.Response()
			slog.Info("request",
				"time", start.Format(time.RFC3339),
				"method", req.Method,
				"uri", req.URL.RequestURI(),
				"status", res.Status,
				"latency", time.Since(start),
				"user_agent", req.UserAgent(),
				"remote_ip", c.RealIP(),
			)
			return err
		}
	})
	e.Use(middleware.Recover())
	e.Pre(middleware.RemoveTrailingSlash())

	e.Validator = &genericValidator{Validator: validator.New()}

	coreService := core.NewCoreService(databaseService, defaultResourcePath, &cfg.Persistence.Cookies, &cfg.Persistence.Media)

	defaultPortStr := strconv.Itoa(cfg.Port)
	apiService := api.NewAPIService(coreService, defaultPortStr)
	apiService.SetAPIRoutes(e)

	uiService := ui.NewUIService(coreService)
	uiService.SetUIRoutes(e)

	// start server
	port := strconv.Itoa(cfg.Port)
	slog.Info("starting server")
	slog.Info(fmt.Sprintf("UI available at http://localhost:%s/%s", port, ui.MainPageName))
	slog.Info(fmt.Sprintf("Explore all feeds via API at http://localhost:%s/%s ", port, api.FeedsPath))
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", port)))
}

type genericValidator struct {
	Validator *validator.Validate
}

func (gv *genericValidator) Validate(i interface{}) error {
	if err := gv.Validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("received invalid request body: %v", err))
	}
	return nil
}
