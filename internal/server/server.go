package server

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jo-hoe/video-to-podcast-service/internal/core"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/common"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"
	"github.com/jo-hoe/video-to-podcast-service/internal/server/api"
	"github.com/jo-hoe/video-to-podcast-service/internal/server/ui"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const defaultPort = "8080"

var defaultResourcePath string

func StartServer(databaseService database.DatabaseService, resourcePath string) {
	defaultResourcePath = resourcePath

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Pre(middleware.RemoveTrailingSlash())

	e.Validator = &genericValidator{Validator: validator.New()}

	coreService := core.NewCoreService(databaseService, defaultResourcePath)

	apiService := api.NewAPIService(coreService, defaultPort)
	apiService.SetAPIRoutes(e)

	uiService := ui.NewUIService(coreService)
	uiService.SetUIRoutes(e)

	// start server
	port := common.ValueOrDefault(os.Getenv("PORT"), "8080")
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
