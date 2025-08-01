package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-playground/validator"
	"github.com/jo-hoe/video-to-podcast-service/internal/config"
	"github.com/jo-hoe/video-to-podcast-service/internal/server/ui"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	cfg := config.LoadUIConfig()
	startUIServer(cfg)
}

func startUIServer(cfg *config.UIConfig) {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Pre(middleware.RemoveTrailingSlash())

	e.Validator = &genericValidator{Validator: validator.New()}

	// Create UI service with API client
	apiClient := NewAPIClient(cfg.API.BaseURL, cfg.API.Timeout)
	uiService := ui.NewUIService(apiClient)
	uiService.SetUIRoutes(e)

	// start UI server
	log.Print("starting UI server")
	log.Printf("UI available at http://localhost:%s/%s", cfg.Server.Port, ui.MainPageName)
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", cfg.Server.Port)))
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
