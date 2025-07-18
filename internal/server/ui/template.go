package ui

import (
	"embed"
	"io"
	"text/template"

	"github.com/labstack/echo/v4"
)

const viewsPattern = "views/*.html"

var (
	//go:embed views/*.html
	templateFS embed.FS
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
