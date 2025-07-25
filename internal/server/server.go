package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/jo-hoe/video-to-podcast-service/internal/config"
	"github.com/jo-hoe/video-to-podcast-service/internal/core"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"
	"github.com/jo-hoe/video-to-podcast-service/internal/server/api"
	"github.com/jo-hoe/video-to-podcast-service/internal/server/ui"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func StartServer(cfg *config.Config, databaseService database.DatabaseService) {

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Pre(middleware.RemoveTrailingSlash())

	e.Validator = &genericValidator{Validator: validator.New()}

	coreService := core.NewCoreService(databaseService, cfg.Storage.BasePath)

	apiService := api.NewAPIService(coreService, cfg.Server.Port)
	apiService.SetAPIRoutes(e)

	uiService := ui.NewUIService(coreService)
	uiService.SetUIRoutes(e)

	// start server
	log.Print("starting server")
	log.Printf("UI available at http://localhost:%s/%s", cfg.Server.Port, ui.MainPageName)
	log.Printf("Explore all feeds via API at http://localhost:%s/%s ", cfg.Server.Port, api.FeedsPath)
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
