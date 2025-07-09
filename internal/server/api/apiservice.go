package api

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

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
	e.GET(fmt.Sprintf("%s%s", FeedsPath, "/:feedTitle/:podcastItemID/:audioFileName"), service.audioFileHandler)
	e.DELETE(fmt.Sprintf("%s%s", FeedsPath, "/:feedTitle/:podcastItemID"), service.deleteFeedItem)

	// Set probe route
	e.GET("/", service.probeHandler)
}

func (service *APIService) deleteFeedItem(ctx echo.Context) error {
	podcastItemID := ctx.Param("podcastItemID")
	feedTitle := ctx.Param("feedTitle")
	validationError := service.validateItemPathComponents(podcastItemID, feedTitle)
	if validationError != nil {
		return validationError
	}

	err := service.coreService.GetDatabaseService().DeletePodcastItem(podcastItemID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to delete podcast item with ID %s", podcastItemID))
	}

	return ctx.NoContent(http.StatusOK)
}

func (service *APIService) validateItemPathComponents(podcastItemID string, feedTitle string) *echo.HTTPError {
	// validate podcastItemID
	if podcastItemID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "podcastItemID is required")
	}

	databaseService := service.coreService.GetDatabaseService()
	podcastItem, err := databaseService.GetPodcastItemByID(podcastItemID)
	if err != nil || podcastItem == nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("no podcast item found with ID %s", podcastItemID))
	}

	// validate feedTitle
	if feedTitle == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "feedTitle is required")
	}

	feedDirectory, err := service.coreService.GetFeedDirectory(podcastItem.AudioFilePath)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "feed item not found (feed directory error)")
	}

	// Normalize both for comparison (case-insensitive, unescape)
	normFeedTitle, _ := url.PathUnescape(feedTitle)
	normFeedDirectory, _ := url.PathUnescape(feedDirectory)
	if !equalPath(normFeedTitle, normFeedDirectory) {
		return echo.NewHTTPError(http.StatusNotFound, "feed item not found (feed title mismatch)")
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
	decodedFeedTitle, err := service.getPathAttributeValue(ctx, "feedTitle")
	if err != nil {
		return err
	}
	decodedAudioFileName, err := service.getPathAttributeValue(ctx, "audioFileName")
	if err != nil {
		return err
	}

	decodedPodcastItemID, err := service.getPathAttributeValue(ctx, "podcastItemID")
	if err != nil {
		return err
	}
	validationError := service.validateItemPathComponents(decodedPodcastItemID, decodedFeedTitle)
	if validationError != nil {
		return validationError
	}

	podcastItem, err := service.coreService.GetDatabaseService().GetPodcastItemByID(decodedPodcastItemID)
	if err != nil {
		return err
	}

	expectedPath := filepath.Clean(filepath.Join(service.coreService.GetAudioSourceDirectory(), decodedFeedTitle, decodedAudioFileName))
	actualPath := filepath.Clean(podcastItem.AudioFilePath)
	if !equalPath(expectedPath, actualPath) {
		return echo.NewHTTPError(http.StatusNotFound, "audio file not found (path mismatch)")
	}

	return ctx.File(podcastItem.AudioFilePath)

}

// equalPath compares two paths for equality, case-insensitive on Windows, and normalizes separators.
func equalPath(a, b string) bool {
	ca := filepath.Clean(a)
	cb := filepath.Clean(b)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(ca, cb)
	}
	return ca == cb
}

func (*APIService) getPathAttributeValue(ctx echo.Context, attributeName string) (string, error) {
	attributeValue := ctx.Param(attributeName)
	if attributeValue == "" {
		return "", echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("%s is required", attributeName))
	}
	decodedFeedTitle, err := url.PathUnescape(attributeValue)
	if err != nil {
		return "", err
	}
	return decodedFeedTitle, nil
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
