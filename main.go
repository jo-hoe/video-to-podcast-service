package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-playground/validator"
	"github.com/gorilla/feeds"
	"github.com/jo-hoe/go-audio-rss-feeder/app/common"
	"github.com/jo-hoe/go-audio-rss-feeder/app/discovery"
	"github.com/jo-hoe/go-audio-rss-feeder/app/download"
	"github.com/jo-hoe/go-audio-rss-feeder/app/feed"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var defaultResourcePath = ""

const defaultPort = "8080"
const defaultItemPath = "v1/feeds"

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

	e.POST("/v1/addItem", addItemHandler)

	e.GET(defaultItemPath, feedsHandler)
	e.GET(fmt.Sprintf("%s%s", defaultItemPath, "/:feedTitle/rss.xml"), feedHandler)
	e.GET(fmt.Sprintf("%s%s", defaultItemPath, "/:feedTitle/:audioFileName"), audioFileHandler)

	e.GET("/", probeHandler)

	port := common.ValueOrDefault(os.Getenv("PORT"), "8080")

	// start server
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", port)))
}

func feedsHandler(ctx echo.Context) (err error) {
	feeds, err := getFeedService().GetFeeds()
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
	feedTitle := ctx.Param("feedTitle")
	result, err := getFeed(feedTitle)
	if err != nil {
		return err
	}

	return ctx.XML(http.StatusOK, result)
}

func audioFileHandler(ctx echo.Context) (err error) {
	feedTitle := ctx.Param("feedTitle")
	if feedTitle == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "feedTitle is required")
	}
	audioFileName := ctx.Param("audioFileName")
	if audioFileName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "audioFileName is required")
	}

	allAudioFiles, err := discovery.GetAudioFiles(getResourcePath())
	if err != nil {
		return err
	}

	foundFile := ""
	for _, audioFile := range allAudioFiles {
		if audioFile == filepath.Join(getResourcePath(), feedTitle, audioFileName) {
			foundFile = audioFile
			break
		}
	}

	if foundFile == "" {
		return echo.NewHTTPError(http.StatusNotFound, "audio file not found")
	}

	return ctx.File(foundFile)
}

func getFeed(feedTitle string) (result *feeds.RssFeed, err error) {
	if feedTitle == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "feedTitle is required")
	}

	feedItems, err := getFeedService().GetFeeds()
	if err != nil {
		return nil, err
	}

	for _, feed := range feedItems {
		if feed.Title == feedTitle {
			result = feed
			break
		}
	}

	if result == nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "feed not found")
	}

	return result, nil
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
	baseUrl := common.ValueOrDefault(os.Getenv("BASE_URL"), getOutboundIP())
	defaultPort := common.ValueOrDefault(os.Getenv("PORT"), defaultPort)
	audioSourceDirectory := getResourcePath()
	return feed.NewFeedService(audioSourceDirectory, baseUrl, defaultPort, defaultItemPath)
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

func getOutboundIP() string {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    localAddr := conn.LocalAddr().(*net.UDPAddr)

    return localAddr.IP.String()
}