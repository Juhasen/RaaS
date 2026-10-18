package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
)

// Same two lines main.go uses; proves the library exposes request metrics on echo v4.
func TestMetricsEndpoint(t *testing.T) {
	e := echo.New()
	e.Use(echoprometheus.NewMiddleware("listing"))
	e.GET("/metrics", echoprometheus.NewHandler())

	e.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/nope", nil))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "listing_request_duration_seconds") {
		t.Fatalf("unexpected metrics response: %d", rec.Code)
	}
}
