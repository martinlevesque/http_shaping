package http_shaping_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/martinlevesque/http_shaping"
)

func TestHttpShapingHappyPath(t *testing.T) {
	cfg := http_shaping.CreateConfig()
	cfg.LoopInterval = 5
	cfg.InTrafficLimit = "1KiB"
	cfg.OutTrafficLimit = "1KiB"
	cfg.ConsiderLimits = true

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	handler, err := http_shaping.New(ctx, next, cfg, "demo-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	if recorder.Code != 200 {
		t.Errorf("Invalid status code")
	}
}

func TestHttpShapingWithLowLimit(t *testing.T) {
	cfg := http_shaping.CreateConfig()
	cfg.LoopInterval = 5
	cfg.InTrafficLimit = "0KiB"
	cfg.OutTrafficLimit = "1KiB"
	cfg.ConsiderLimits = true

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	handler, err := http_shaping.New(ctx, next, cfg, "demo-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	if recorder.Code != 429 {
		t.Errorf("Invalid status code with low limit")
	}
}

func TestHttpShapingWithLowLimitButShouldIgnore(t *testing.T) {
	cfg := http_shaping.CreateConfig()
	cfg.LoopInterval = 5
	cfg.InTrafficLimit = "0KiB"
	cfg.OutTrafficLimit = "1KiB"
	cfg.ConsiderLimits = false

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	handler, err := http_shaping.New(ctx, next, cfg, "demo-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	if recorder.Code != 200 {
		t.Errorf("Invalid status code")
	}
}

