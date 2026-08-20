package api

import (
	"net/http"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
)

type testValidator struct {
	v *validator.Validate
}

func newRequestValidator() echo.Validator {
	return &testValidator{v: validator.New()}
}

func (tv *testValidator) Validate(i interface{}) error {
	if err := tv.v.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request data")
	}
	return nil
}
