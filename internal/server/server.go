package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

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
	e.Use(middleware.Logger())
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
	log.Print("starting server")
	log.Printf("UI available at http://localhost:%s/%s", port, ui.MainPageName)
	log.Printf("Explore all feeds via API at http://localhost:%s/%s ", port, api.FeedsPath)
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
