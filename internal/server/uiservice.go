package server

import (
	"net/http"

	"github.com/jo-hoe/video-to-podcast-service/internal/core"
	"github.com/labstack/echo/v4"
)

const mainPageName = "index.html"

type DownloadItem struct {
	URL string `json:"url" validate:"required"`
}

type DownloadItems struct {
	URLS []string `json:"urls" validate:"required"`
}

type UIService struct {
	coreservice *core.CoreService
}

func NewUIService(coreservice *core.CoreService) *UIService {
	return &UIService{
		coreservice: coreservice,
	}
}

func (service *UIService) setUIRoutes(e *echo.Echo) {
	// Set UI routes
	e.GET(mainPageName, service.indexHandler)
	e.POST("/htmx/addItem", service.htmxAddItemHandler) // new HTMX endpoint
}

func (service *UIService) indexHandler(ctx echo.Context) (err error) {
	return ctx.Render(http.StatusOK, "index", nil)
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
