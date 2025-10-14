package api

import (
	"fmt"
	"log"
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
	e.GET(fmt.Sprintf("%s%s", FeedsPath, "/:feedTitle/:audioFileName"), service.audioFileHandler)
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

	podcastItem, err := service.coreService.GetDatabaseService().GetPodcastItemByID(podcastItemID)
	if err != nil {
		log.Printf("failed to retrieve podcast item with ID %s: %v", podcastItemID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete podcast item.")
	}
	err = service.coreService.GetDatabaseService().DeletePodcastItem(podcastItem.ID)
	if err != nil {
		log.Printf("failed to delete podcast item with ID %s: %v", podcastItemID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete podcast item.")
	}

	// Remove the audio file if it exists
	err = os.Remove(podcastItem.AudioFilePath)
	if err != nil && !os.IsNotExist(err) {
		log.Printf("failed to delete audio file for podcast item with ID %s: %v", podcastItemID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete podcast item.")
	}

	// check if directory is empty and remove it if so
	feedDirectory := filepath.Join(service.coreService.GetAudioSourceDirectory(), feedTitle)
	if _, err := os.Stat(feedDirectory); err == nil {
		files, err := os.ReadDir(feedDirectory)
		if err != nil {
			log.Printf("failed to read directory %s: %v", feedDirectory, err)
			return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
		}
		if len(files) == 0 {
			err = os.Remove(feedDirectory)
			if err != nil {
				log.Printf("failed to remove empty feed directory %s: %v", feedDirectory, err)
				return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
			}
			log.Printf("removed empty feed directory: %s", feedDirectory)
		} else {
			log.Printf("feed directory %s is not empty, skipping removal", feedDirectory)
		}
	} else if !os.IsNotExist(err) {
		log.Printf("failed to check if feed directory %s exists: %v", feedDirectory, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	return ctx.NoContent(http.StatusOK)
}

func (service *APIService) validateItemPathComponents(podcastItemID string, feedTitle string) *echo.HTTPError {
	// validate podcastItemID
	if podcastItemID == "" {
		log.Printf("podcastItemID is required for validation")
		return echo.NewHTTPError(http.StatusBadRequest, "podcastItemID is required")
	}

	databaseService := service.coreService.GetDatabaseService()
	podcastItem, err := databaseService.GetPodcastItemByID(podcastItemID)
	if err != nil || podcastItem == nil {
		log.Printf("no podcast item found with ID %s", podcastItemID)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid podcast item")
	}

	// validate feedTitle
	if feedTitle == "" {
		log.Printf("feedTitle is required for validation")
		return echo.NewHTTPError(http.StatusBadRequest, "feedTitle is required")
	}

	feedDirectory, err := service.coreService.GetFeedDirectory(podcastItem.AudioFilePath)
	if err != nil {
		log.Printf("feed item not found (feed directory error) for podcast item ID %s", podcastItemID)
		return echo.NewHTTPError(http.StatusNotFound, "feed item not found")
	}

	// Normalize both for comparison (case-insensitive, unescape)
	normFeedTitle, _ := url.PathUnescape(feedTitle)
	normFeedDirectory, _ := url.PathUnescape(feedDirectory)
	if !equalPath(normFeedTitle, normFeedDirectory) {
		log.Printf("feed item not found (feed title mismatch) for podcast item ID %s", podcastItemID)
		return echo.NewHTTPError(http.StatusNotFound, "feed item not found")
	}

	return nil
}

func (service *APIService) feedsHandler(ctx echo.Context) (err error) {
	feeds, err := service.getFeedService().GetFeeds(ctx.Request().Host)
	if err != nil {
		log.Printf("failed to get feeds: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get feeds")
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
		log.Printf("failed to bind download items: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err = ctx.Validate(downloadItems); err != nil {
		log.Printf("failed to validate download items: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request data")
	}

	for _, url := range downloadItems.URLS {
		err = service.coreService.DownloadItemsHandler(url)
		if err != nil {
			log.Printf("failed to handle download for url %s: %v", url, err)
			return echo.NewHTTPError(http.StatusBadRequest, "unsupported URL")
		}
	}

	return ctx.NoContent(http.StatusOK)
}

func (service *APIService) feedHandler(ctx echo.Context) (err error) {
	feedTitle := ctx.Param("feedTitle")
	result, err := service.getFeed(ctx.Request().Host, feedTitle)
	if err != nil {
		log.Printf("failed to get feed for title %s: %v", feedTitle, err)
		return echo.NewHTTPError(http.StatusNotFound, "feed not found")
	}

	rss, err := result.ToRss()
	if err != nil {
		log.Printf("failed to generate RSS for feed %s: %v", feedTitle, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate RSS")
	}
	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationXMLCharsetUTF8)
	_, err = ctx.Response().Writer.Write([]byte(rss))
	return err
}

func (service *APIService) audioFileHandler(ctx echo.Context) (err error) {
	decodedFeedTitle, err := service.getPathAttributeValue(ctx, "feedTitle")
	if err != nil {
		log.Printf("failed to get feedTitle from path: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid feed title")
	}
	decodedAudioFileName, err := service.getPathAttributeValue(ctx, "audioFileName")
	if err != nil {
		log.Printf("failed to get audioFileName from path: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid audio file name")
	}

	expectedPath := filepath.Clean(filepath.Join(service.coreService.GetAudioSourceDirectory(), decodedFeedTitle, decodedAudioFileName))
	podcastItems, err := service.coreService.GetDatabaseService().GetAllPodcastItems()
	if err != nil {
		log.Printf("failed to retrieve podcast items: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to retrieve podcast items")
	}

	foundItem := false
	// loop through all podcast items to find the one with the expected audio file path
	// this can be optimized by introducing a dedicated method in the database service to find a podcast item by its audio file path
	for _, item := range podcastItems {
		if equalPath(item.AudioFilePath, expectedPath) {
			foundItem = true
			break
		}
	}
	if !foundItem {
		log.Printf("audio file not found for feed %s and audio file %s", decodedFeedTitle, decodedAudioFileName)
		return echo.NewHTTPError(http.StatusNotFound, "audio file not found")
	}

	return ctx.File(expectedPath)
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

// getPathAttributeValue returns a decoded path parameter or an error if missing/invalid.
func (*APIService) getPathAttributeValue(ctx echo.Context, attributeName string) (string, error) {
	value := ctx.Param(attributeName)
	if value == "" {
		return "", echo.NewHTTPError(http.StatusBadRequest, attributeName+" is required")
	}
	return url.PathUnescape(value)
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
