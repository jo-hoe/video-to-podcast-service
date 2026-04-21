package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jo-hoe/gofeedx"
	"github.com/jo-hoe/video-to-podcast-service/internal/core"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/feed"
	"github.com/jo-hoe/video-to-podcast-service/internal/server/requestutil"
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

	// Health endpoint for Kubernetes probes
	e.GET(HealthPath, service.healthHandler)

	// Set probe route
	e.GET(ProbePath, service.probeHandler)
}

func (service *APIService) deleteFeedItem(ctx echo.Context) error {
	podcastItemID := ctx.Param("podcastItemID")
	feedTitle := ctx.Param("feedTitle")
	validationError := service.validateItemPathComponents(podcastItemID, feedTitle)
	if validationError != nil {
		return validationError
	}

	err := service.coreService.DeletePodcastItem(podcastItemID)
	if err != nil {
		slog.Error("failed to delete podcast item", "podcastItemID", podcastItemID, "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete podcast item.")
	}

	return ctx.NoContent(http.StatusOK)
}

func (service *APIService) validateItemPathComponents(podcastItemID string, feedTitle string) *echo.HTTPError {
	// validate podcastItemID
	if podcastItemID == "" {
		slog.Warn("podcastItemID is required for validation")
		return echo.NewHTTPError(http.StatusBadRequest, "podcastItemID is required")
	}

	databaseService := service.coreService.GetDatabaseService()
	podcastItem, err := databaseService.GetPodcastItemByID(podcastItemID)
	if err != nil || podcastItem == nil {
		slog.Warn("no podcast item found", "podcastItemID", podcastItemID)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid podcast item")
	}

	// validate feedTitle
	if feedTitle == "" {
		slog.Warn("feedTitle is required for validation")
		return echo.NewHTTPError(http.StatusBadRequest, "feedTitle is required")
	}

	feedDirectory, err := service.coreService.GetFeedDirectory(podcastItem.AudioFilePath)
	if err != nil {
		slog.Warn("feed item not found (feed directory error)", "podcastItemID", podcastItemID)
		return echo.NewHTTPError(http.StatusNotFound, "feed item not found")
	}

	// Normalize both for comparison (case-insensitive, unescape)
	normFeedTitle, _ := url.PathUnescape(feedTitle)
	normFeedDirectory, _ := url.PathUnescape(feedDirectory)
	if !equalPath(normFeedTitle, normFeedDirectory) {
		slog.Warn("feed item not found (feed title mismatch)", "podcastItemID", podcastItemID)
		return echo.NewHTTPError(http.StatusNotFound, "feed item not found")
	}

	return nil
}

func (service *APIService) feedsHandler(ctx echo.Context) (err error) {
	baseURL := requestutil.BaseURL(ctx)
	feeds, err := service.getFeedService().GetFeeds(baseURL)
	if err != nil {
		slog.Error("failed to get feeds", "err", err)
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
		slog.Error("failed to bind download items", "err", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err = ctx.Validate(downloadItems); err != nil {
		slog.Error("failed to validate download items", "err", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request data")
	}

	for _, url := range downloadItems.URLS {
		err = service.coreService.DownloadItemsHandler(url)
		if err != nil {
			slog.Error("failed to handle download", "url", url, "err", err)
			return echo.NewHTTPError(http.StatusBadRequest, "unsupported URL")
		}
	}

	return ctx.NoContent(http.StatusOK)
}

func (service *APIService) feedHandler(ctx echo.Context) (err error) {
	feedTitle := ctx.Param("feedTitle")
	baseURL := requestutil.BaseURL(ctx)
	result, err := service.getFeed(baseURL, feedTitle)
	if err != nil {
		slog.Error("failed to get feed", "feedTitle", feedTitle, "err", err)
		return echo.NewHTTPError(http.StatusNotFound, "feed not found")
	}

	psp, err := gofeedx.ToPSP(result)
	if err != nil {
		slog.Error("failed to generate PSP", "feedTitle", feedTitle, "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate PSP")
	}
	ctx.Response().Header().Set(echo.HeaderContentType, "application/rss+xml; charset=utf-8")
	_, err = ctx.Response().Writer.Write([]byte(psp))
	return err
}

func (service *APIService) audioFileHandler(ctx echo.Context) (err error) {
	decodedFeedTitle, err := service.getPathAttributeValue(ctx, "feedTitle")
	if err != nil {
		slog.Error("failed to get feedTitle from path", "err", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid feed title")
	}
	decodedAudioFileName, err := service.getPathAttributeValue(ctx, "audioFileName")
	if err != nil {
		slog.Error("failed to get audioFileName from path", "err", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid audio file name")
	}

	expectedPath := filepath.Clean(filepath.Join(service.coreService.GetAudioSourceDirectory(), decodedFeedTitle, decodedAudioFileName))
	podcastItems, err := service.coreService.GetDatabaseService().GetAllPodcastItems()
	if err != nil {
		slog.Error("failed to retrieve podcast items", "err", err)
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
		slog.Warn("audio file not found", "feedTitle", decodedFeedTitle, "audioFileName", decodedAudioFileName)
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

func (service *APIService) getFeed(baseURL *url.URL, feedTitle string) (result *gofeedx.Feed, err error) {
	if feedTitle == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "feedTitle is required")
	}

	feedItems, err := service.getFeedService().GetFeeds(baseURL)
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
	return feed.NewFeedService(service.coreService, service.defaultPort, FeedsPath)
}