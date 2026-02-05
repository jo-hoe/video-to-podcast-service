package ui

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"sort"
	"text/template"

	"github.com/jo-hoe/video-to-podcast-service/internal/core"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"
	"github.com/jo-hoe/video-to-podcast-service/internal/server/api"
	"github.com/labstack/echo/v4"
)

const MainPageName = "index.html"

type UIService struct {
	coreservice *core.CoreService
}

type PodcastItemList struct {
	PodcastItems []*database.PodcastItem
	Host         string
	APIPath      string
}

func NewUIService(coreservice *core.CoreService) *UIService {
	return &UIService{
		coreservice: coreservice,
	}
}

func (service *UIService) SetUIRoutes(e *echo.Echo) {
	// Create template with helper functions
	funcMap := template.FuncMap{
		"formatDuration":       formatDuration,
		"getFeedLink":          service.getFeedLink,
		"getFeedTitleFromPath": getFeedTitleFromPath,
	}

	e.Renderer = &Template{
		templates: template.Must(template.New("").Funcs(funcMap).ParseFS(templateFS, viewsPattern)),
	}
	// Set UI routes
	e.GET("/", service.rootRedirectHandler) // Redirect root to index.html
	e.GET(MainPageName, service.indexHandler)
	e.POST("/htmx/addItem", service.htmxAddItemHandler) // new HTMX endpoint
	e.GET("/icon.svg", service.iconHandler)
}

// rootRedirectHandler redirects root path to index.html
func (service *UIService) rootRedirectHandler(ctx echo.Context) error {
	return ctx.Redirect(http.StatusMovedPermanently, "/"+MainPageName)
}

// Helper function to extract feed title from audio file path
func getFeedTitleFromPath(path string) string {
	// Assumes feed title is the parent directory of the audio file
	return filepath.Base(filepath.Dir(path))
}

func (service *UIService) indexHandler(ctx echo.Context) (err error) {
	// Get all podcast items from database
	podcastItems, err := service.coreservice.GetDatabaseService().GetAllPodcastItems()
	if err != nil {
		// Log error but don't fail the page load
		podcastItems = []*database.PodcastItem{}
	}

	// Sort by UpdatedAt in descending order (most recent first)
	sort.Slice(podcastItems, func(i, j int) bool {
		return podcastItems[i].UpdatedAt.After(podcastItems[j].UpdatedAt)
	})

	// reduce the number of results to 128 items
	if len(podcastItems) > 128 {
		podcastItems = podcastItems[:128]
	}

	// Get host and API path from the request
	host := ctx.Request().Host
	apiPath := "api" // You might want to make this configurable

	data := PodcastItemList{
		PodcastItems: podcastItems,
		Host:         host,
		APIPath:      apiPath,
	}

	return ctx.Render(http.StatusOK, "index", data)
}

// New handler for HTMX single URL form
func (service *UIService) htmxAddItemHandler(ctx echo.Context) error {
	type SingleUrl struct {
		URL string `json:"url" form:"url" validate:"required"`
	}
	var req SingleUrl
	if err := ctx.Bind(&req); err != nil || req.URL == "" {
		return ctx.HTML(http.StatusBadRequest, "<span style='color:red'>Invalid or missing URL.</span>")
	}
	if err := service.coreservice.DownloadItemsHandler(req.URL); err != nil {
		return ctx.HTML(http.StatusInternalServerError, "<span style='color:red'>Failed to process: "+req.URL+"</span>")
	}
	return ctx.HTML(http.StatusOK, "<span style='color:green'>Submitted successfully!</span>")
}

 // Icon handler to serve the embedded favicon
func (service *UIService) iconHandler(ctx echo.Context) error {
	file, err := templateFS.Open("views/icon.svg")
	if err != nil {
		return ctx.NoContent(http.StatusNotFound)
	}
	defer file.Close()

	ctx.Response().Header().Set(echo.HeaderContentType, "image/svg+xml")
	_, err = io.Copy(ctx.Response().Writer, file)
	return err
}

// Helper function to format duration from milliseconds to HH:MM:SS or MM:SS
func formatDuration(milliseconds int64) string {
	totalSeconds := milliseconds / 1000
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

// Helper function to generate feed link for a podcast item
func (service *UIService) getFeedLink(host, filePath string) string {
	return service.coreservice.GetLinkToFeed(host, api.FeedsPath, filePath)
}
