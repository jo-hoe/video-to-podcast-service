package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-playground/validator"
	"github.com/jo-hoe/go-audio-rss-feeder/app/common"
	"github.com/jo-hoe/go-audio-rss-feeder/app/download"
	"github.com/jo-hoe/go-audio-rss-feeder/app/feed"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var defaultResourcePath = ""

const defaultPort = "8080"

func getResourcePath() string {
	if defaultResourcePath != "" {
		return defaultResourcePath
	}

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	defaultResourcePath = common.ValueOrDefault(os.Getenv("BASE_PATH"), filepath.Join(exPath, "resources"))
	return defaultResourcePath
}

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Validator = &genericValidator{Validator: validator.New()}

	e.GET("/v1/feeds", feedsHandler)
	e.GET("/v1/feeds/:feedName", feedHandler)
	e.POST("/v1/addItem", addItemHandler)
	e.GET("/", probeHandler)

	port := common.ValueOrDefault(os.Getenv("PORT"), "8080")

	// start server
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", port)))
}

func feedsHandler(ctx echo.Context) (err error) {
	baseUrl := common.ValueOrDefault(os.Getenv("BASE_URL"), "127.0.0.1")
	defaultPort := common.ValueOrDefault(os.Getenv("PORT"), defaultPort)
	audioSourceDirectory := getResourcePath()
	feeds, err := feed.NewFeedService(audioSourceDirectory, baseUrl, defaultPort).GetFeeds()
	if err != nil {
		return err
	}

	result := make([]string, 0)
	for _, feed := range feeds {
		result = append(result, feed.Link)
	}

	return ctx.JSON(http.StatusOK, result)
}

func feedHandler(ctx echo.Context) (err error) {
	baseUrl := common.ValueOrDefault(os.Getenv("BASE_URL"), "127.0.0.1")
	defaultPort := common.ValueOrDefault(os.Getenv("PORT"), defaultPort)
	audioSourceDirectory := getResourcePath()
	feeds, err := feed.NewFeedService(audioSourceDirectory, baseUrl, defaultPort).GetFeeds()
	if err != nil {
		return err
	}

	result := make([]string, 0)
	for _, feed := range feeds {
		result = append(result, feed.Link)
	}

	return ctx.JSON(http.StatusOK, result)
}

type DownloadItem struct {
	URL string `json:"url" validate:"required"`
}

func addItemHandler(ctx echo.Context) (err error) {
	downloadItem := new(DownloadItem)
	if err = ctx.Bind(downloadItem); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err = ctx.Validate(downloadItem); err != nil {
		return err
	}

	downloader := download.YoutubeAudioDownloader{}
	audioSourceDirectory := getResourcePath()
	_, err = downloader.Download(downloadItem.URL, audioSourceDirectory)
	if err != nil {
		return err
	}

	return ctx.NoContent(http.StatusOK)
}

func probeHandler(ctx echo.Context) (err error) {
	return ctx.NoContent(http.StatusOK)
}

func getFeedService() *feed.FeedService {
	baseUrl := common.ValueOrDefault(os.Getenv("BASE_URL"), "127.0.0.1")
	defaultPort := common.ValueOrDefault(os.Getenv("PORT"), defaultPort)
	audioSourceDirectory := getResourcePath()
	return feed.NewFeedService(audioSourceDirectory, baseUrl, defaultPort)
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
