package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

const mainPageName = "index.html"

func setUIRoutes(e *echo.Echo) {
	// Set UI routes
	e.GET(mainPageName, indexHandler)
	e.POST("/htmx/addItem", htmxAddItemHandler) // new HTMX endpoint
}

func indexHandler(ctx echo.Context) (err error) {
	return ctx.Render(http.StatusOK, "index", nil)
}

type DownloadItems struct {
	URLS []string `json:"urls" validate:"required"`
}

func addItemsHandler(ctx echo.Context) (err error) {
	downloadItems := new(DownloadItems)
	if err = ctx.Bind(downloadItems); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err = ctx.Validate(downloadItems); err != nil {
		return err
	}

	for _, url := range downloadItems.URLS {
		err = downloadItemsHandler(url)
		if err != nil {
			return err
		}
	}

	return ctx.NoContent(http.StatusOK)
}

// New handler for HTMX single URL form
func htmxAddItemHandler(ctx echo.Context) error {
	type SingleUrl struct {
		URL string `json:"url" form:"url" validate:"required"`
	}
	var req SingleUrl
	if err := ctx.Bind(&req); err != nil || req.URL == "" {
		return ctx.HTML(http.StatusBadRequest, "<span style='color:red'>Invalid or missing URL.</span>")
	}
	if err := downloadItemsHandler(req.URL); err != nil {
		return ctx.HTML(http.StatusInternalServerError, "<span style='color:red'>Failed to process: "+req.URL+"</span>")
	}
	return ctx.HTML(http.StatusOK, "<span style='color:green'>Submitted successfully!</span>")
}
