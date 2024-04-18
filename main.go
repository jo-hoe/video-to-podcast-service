package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/v1/feed", feedHandler)
	e.POST("/v1/addItem", addItemHandler)
	e.GET("/", probeHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// start server
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", port)))
}

func feedHandler(ctx echo.Context) (err error) {
	NewFeedProvider("","","","","","","").GetFeed()

	return nil
}

type DownloadItem struct {
	URL string `json:"url" validate:"required"`
}

func addItemHandler(ctx echo.Context) (err error) {
	return nil
}

func probeHandler(ctx echo.Context) (err error) {
	return ctx.NoContent(http.StatusOK)
}
