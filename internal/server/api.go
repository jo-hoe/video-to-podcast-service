package server

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/go-playground/validator"
	"github.com/gorilla/feeds"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/common"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/download"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/feed"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/filemanagement"
	"github.com/labstack/echo/v4"
)

const apiVersion = "v1/"
const feedsPath = apiVersion + "feeds"
const addItemPaths = apiVersion + "addItems"

func setAPIRoutes(e *echo.Echo) {
	// API routes
	e.POST(addItemPaths, addItemsHandler)
	e.GET(feedsPath, feedsHandler)
	e.GET(fmt.Sprintf("%s%s", feedsPath, "/:feedTitle/rss.xml"), feedHandler)
	e.GET(fmt.Sprintf("%s%s", feedsPath, "/:feedTitle/:audioFileName"), audioFileHandler)

	// Set probe route
	e.GET("/", probeHandler)
}

func feedsHandler(ctx echo.Context) (err error) {
	feeds, err := getFeedService().GetFeeds(ctx.Request().Host)
	if err != nil {
		return err
	}

	result := make([]string, 0)
	for _, feed := range feeds {
		result = append(result, feed.Link.Href)
	}

	return ctx.JSON(http.StatusOK, result)
}

func feedHandler(ctx echo.Context) (err error) {
	feedTitle := ctx.Param("feedTitle")
	result, err := getFeed(ctx.Request().Host, feedTitle)
	if err != nil {
		return err
	}

	rss, err := result.ToRss()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to generate RSS: %v", err))
	}
	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationXMLCharsetUTF8)
	_, err = ctx.Response().Writer.Write([]byte(rss))
	return err
}

func audioFileHandler(ctx echo.Context) (err error) {
	feedTitle := ctx.Param("feedTitle")
	if feedTitle == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "feedTitle is required")
	}
	decodedFeedTitle, err := url.QueryUnescape(feedTitle)
	if err != nil {
		return err
	}
	audioFileName := ctx.Param("audioFileName")
	if audioFileName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "audioFileName is required")
	}
	decodedAudioFileName, err := url.QueryUnescape(audioFileName)
	if err != nil {
		return err
	}

	allAudioFiles, err := filemanagement.GetAudioFiles(defaultResourcePath)
	if err != nil {
		return err
	}

	foundFile := ""
	for _, audioFile := range allAudioFiles {
		if audioFile == filepath.Join(defaultResourcePath, decodedFeedTitle, decodedAudioFileName) {
			foundFile = audioFile
			break
		}
	}

	if foundFile == "" {
		return echo.NewHTTPError(http.StatusNotFound, "audio file not found")
	}

	return ctx.File(foundFile)
}

func getFeed(host, feedTitle string) (result *feeds.Feed, err error) {
	if feedTitle == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "feedTitle is required")
	}

	feedItems, err := getFeedService().GetFeeds(host)
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

func downloadItemsHandler(url string) (err error) {
	downloader, err := download.GetVideoDownloader(url)
	if err != nil {
		return err
	}
	audioSourceDirectory := defaultResourcePath
	if !downloader.IsVideoAvailable(url) {
		return fmt.Errorf("video %s is not available", url)
	}
	log.Printf("downloading '%s'", url)

	go func() {
		maxErrorCount := 4
		errorCount := 0

		for errorCount < maxErrorCount {
			_, err := downloader.Download(url, audioSourceDirectory)
			if err != nil {
				log.Printf("failed to download '%s': %v", url, err)
				errorCount++
			} else {
				break
			}
		}
	}()

	return nil
}

func probeHandler(ctx echo.Context) (err error) {
	return ctx.NoContent(http.StatusOK)
}

func getFeedService() *feed.FeedService {
	defaultPort := common.ValueOrDefault(os.Getenv("PORT"), defaultPort)
	audioSourceDirectory := defaultResourcePath

	return feed.NewFeedService(audioSourceDirectory, defaultPort, feedsPath)
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
