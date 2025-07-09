package api

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/gorilla/feeds"
	"github.com/jo-hoe/video-to-podcast-service/internal/core"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/common"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/feed"
	"github.com/labstack/echo/v4"
)

const (
	apiVersion   = "v1/"
	addItemPaths = apiVersion + "addItems"

	FeedsPath = apiVersion + "feeds"
)

type APIService struct {
	coreService *core.CoreService
	defaultPort string
}

type DownloadItem struct {
	URL string `json:"url" validate:"required"`
}

type DownloadItems struct {
	URLS []string `json:"urls" validate:"required"`
}

func NewAPIService(coreservice *core.CoreService, defaultPort string) *APIService {
	return &APIService{
		coreService: coreservice,
		defaultPort: defaultPort,
	}
}

func (service *APIService) SetAPIRoutes(e *echo.Echo) {
	// API routes
	e.POST(addItemPaths, service.addItemsHandler)
	e.GET(FeedsPath, service.feedsHandler)
	e.GET(fmt.Sprintf("%s%s", FeedsPath, "/:feedTitle/rss.xml"), service.feedHandler)
	e.GET(fmt.Sprintf("%s%s", FeedsPath, "/:feedTitle/:audioFileName"), service.audioFileHandler)
	e.DELETE(fmt.Sprintf("%s%s", FeedsPath, "/:feedTitle/:podcastItemID"), service.feedHandler)

	// Set probe route
	e.GET("/", service.probeHandler)
}

func (service *APIService) deleteFeedItem(ctx echo.Context) error {
	// validate podcastItemID
	podcastItemID := ctx.Param("podcastItemID")
	if podcastItemID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "itemItemID is required")
	}

	databaseService := service.coreService.GetDatabaseService()
	podcastItem, err := databaseService.GetPodcastItemByID(podcastItemID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("no podcast item found with ID %s", podcastItemID))
	}

	// validate feedTitle
	feedTitle := ctx.Param("feedTitle")
	if feedTitle == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "feedTitle is required")
	}
	feedDirectory, err := service.coreService.GetFeedDirectory(podcastItem.AudioFilePath)
	if err != nil || podcastItem == nil || feedTitle != feedDirectory {
		return echo.NewHTTPError(http.StatusNotFound, "feed item not found")
	}

	err = databaseService.DeletePodcastItem(podcastItemID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to delete podcast item with ID %s", podcastItemID))
	}

	return ctx.NoContent(http.StatusOK)
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
		err = service.coreService.DownloadItemsHandler(url)
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

	pocastItems, err := service.coreService.GetDatabaseService().GetAllPodcastItems()
	if err != nil {
		return err
	}

	foundFile := ""
	for _, audioFile := range pocastItems {
		if audioFile.AudioFilePath == filepath.Join(service.coreService.GetAudioSourceDirectory(), decodedFeedTitle, decodedAudioFileName) {
			foundFile = audioFile.AudioFilePath
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
	port := common.ValueOrDefault(os.Getenv("PORT"), service.defaultPort)

	return feed.NewFeedService(service.coreService, port, FeedsPath)
}
