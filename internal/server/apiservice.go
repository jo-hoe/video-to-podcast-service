package server

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/go-playground/validator"
	"github.com/gorilla/feeds"
	"github.com/jo-hoe/video-to-podcast-service/internal/core"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/common"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/feed"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/filemanagement"
	"github.com/labstack/echo/v4"
)

const (
	apiVersion   = "v1/"
	feedsPath    = apiVersion + "feeds"
	addItemPaths = apiVersion + "addItems"
)

type APIService struct {
	coreservice *core.CoreService
}

func NewAPIService(coreservice *core.CoreService) *APIService {
	return &APIService{
		coreservice: coreservice,
	}
}

func (service *APIService) setAPIRoutes(e *echo.Echo) {
	// API routes
	e.POST(addItemPaths, service.addItemsHandler)
	e.GET(feedsPath, service.feedsHandler)
	e.GET(fmt.Sprintf("%s%s", feedsPath, "/:feedTitle/rss.xml"), service.feedHandler)
	e.GET(fmt.Sprintf("%s%s", feedsPath, "/:feedTitle/:audioFileName"), service.audioFileHandler)

	// Set probe route
	e.GET("/", service.probeHandler)
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

func (service *APIService) feedsHandler(ctx echo.Context) (err error) {
	feeds, err := service.getFeedService().GetFeeds(ctx.Request().Host)
	if err != nil {
		return err
	}

	result := make([]string, 0)
	for _, feed := range feeds {
		result = append(result, feed.Link.Href)
	}

	return ctx.JSON(http.StatusOK, result)
}

func (service *APIService) addItemsHandler(ctx echo.Context) (err error) {
	downloadItems := new(DownloadItems)
	if err = ctx.Bind(downloadItems); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err = ctx.Validate(downloadItems); err != nil {
		return err
	}

	for _, url := range downloadItems.URLS {
		err = service.coreservice.DownloadItemsHandler(url)
		if err != nil {
			return err
		}
	}

	return ctx.NoContent(http.StatusOK)
}

func (service *APIService) feedHandler(ctx echo.Context) (err error) {
	feedTitle := ctx.Param("feedTitle")
	result, err := service.getFeed(ctx.Request().Host, feedTitle)
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

func (service *APIService) audioFileHandler(ctx echo.Context) (err error) {
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

func (service *APIService) getFeed(host, feedTitle string) (result *feeds.Feed, err error) {
	if feedTitle == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "feedTitle is required")
	}

	feedItems, err := service.getFeedService().GetFeeds(host)
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

func (service *APIService) probeHandler(ctx echo.Context) (err error) {
	return ctx.NoContent(http.StatusOK)
}

func (service *APIService) getFeedService() *feed.FeedService {
	defaultPort := common.ValueOrDefault(os.Getenv("PORT"), defaultPort)

	return feed.NewFeedService(service.coreservice, defaultPort, feedsPath)
}
