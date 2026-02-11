package server

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
)

type sampleReq struct {
	URL string `validate:"required,url"`
	ID  int    `validate:"min=1"`
}

func TestGenericValidator_ValidStruct(t *testing.T) {
	gv := &genericValidator{Validator: validator.New()}
	req := sampleReq{
		URL: "http://example.com",
		ID:  1,
	}
	if err := gv.Validate(req); err != nil {
		t.Fatalf("expected no error for valid struct, got %v", err)
	}
}

func TestGenericValidator_InvalidStruct_ReturnsError(t *testing.T) {
	gv := &genericValidator{Validator: validator.New()}
	req := sampleReq{
		URL: "",
		ID:  0,
	}
	if err := gv.Validate(req); err == nil {
		t.Fatalf("expected error for invalid struct")
	}
}

func TestGenericValidator_InvalidStruct_ReturnsHTTPError(t *testing.T) {
	gv := &genericValidator{Validator: validator.New()}
	req := sampleReq{
		URL: "",
		ID:  0,
	}
	err := gv.Validate(req)
	if err == nil {
		t.Fatalf("expected error for invalid struct")
	}
	if _, ok := err.(*echo.HTTPError); !ok {
		t.Fatalf("expected *echo.HTTPError, got %#v", err)
	}
}

func TestGenericValidator_InvalidStruct_StatusBadRequest(t *testing.T) {
	gv := &genericValidator{Validator: validator.New()}
	req := sampleReq{
		URL: "",
		ID:  0,
	}
	err := gv.Validate(req)
	if err == nil {
		t.Fatalf("expected error for invalid struct")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %#v", err)
	}
	if httpErr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 BadRequest, got %d", httpErr.Code)
	}
}

func TestGenericValidator_InvalidStruct_MessageContainsInvalidBody(t *testing.T) {
	gv := &genericValidator{Validator: validator.New()}
	req := sampleReq{
		URL: "",
		ID:  0,
	}
	err := gv.Validate(req)
	if err == nil {
		t.Fatalf("expected error for invalid struct")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %#v", err)
	}
	formatted := strings.ToLower(strings.TrimSpace(strings.ReplaceAll(fmt.Sprintf("%v", httpErr.Message), "\n", " ")))
	if !strings.Contains(formatted, "received invalid request body") {
		t.Fatalf("expected error message to contain 'received invalid request body', got %q", formatted)
	}
}
