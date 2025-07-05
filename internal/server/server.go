package server

import (
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/jo-hoe/video-to-podcast-service/internal/core"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/common"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"

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
	e.Renderer = &Template{
		templates: template.Must(template.ParseFS(templateFS, viewsPattern)),
	}

	coreService := core.NewCoreService(databaseService, defaultResourcePath)

	apiService := NewAPIService(coreService)
	apiService.setAPIRoutes(e)

	uiService := NewUIService(coreService)
	uiService.setUIRoutes(e)

	// start server
	port := common.ValueOrDefault(os.Getenv("PORT"), "8080")
	log.Print("starting server")
	log.Printf("UI available at http://localhost:%s/%s", port, mainPageName)
	log.Printf("Explore all feeds via API at http://localhost:%s/%s ", port, feedsPath)
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", port)))
}
