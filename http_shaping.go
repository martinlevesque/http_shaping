// Package plugindemo a demo plugin.
package http_shaping

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"text/template"
	"time"
	"strconv"
	"github.com/martinlevesque/http_shaping/bytefmt"
)

// Config the plugin configuration.
type Config struct {
	Headers map[string]string `json:"headers,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		Headers: make(map[string]string),
	}
}

// Demo a Demo plugin.
type Demo struct {
	next     http.Handler
	headers  map[string]string
	name     string
	template *template.Template
	loopBeginTimestamp int64
	loopSumInBytes int64
	loopSumOutBytes int64
}

// New created a new Demo plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	fmt.Print("Initiating http_shaping plugin ", name, "\n")
	what := bytefmt.ToBytes("10.5GiB")
	fmt.Print("what ", what, "\n")

	if len(config.Headers) == 0 {
		return nil, fmt.Errorf("headers cannot be empty")
	}

	return &Demo{
		headers:  config.Headers,
		next:     next,
		name:     name,
		template: template.New("demo").Delims("[[", "]]"),
		loopBeginTimestamp: time.Now().Unix(),
		loopSumInBytes: 0,
		loopSumOutBytes: 0,
	}, nil
}

func (a *Demo) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	fmt.Print("Serving http_shaping plugin \n")

	fmt.Print("loopSumInBytes: ", a.loopSumInBytes, "\n")
	fmt.Print("loopSumOutBytes: ", a.loopSumOutBytes, "\n")

	requestBytesStr := req.Header.Get("content-length")

	if requestBytesStr != "" {
		requestBytes, err := strconv.ParseInt(requestBytesStr, 10, 64)

		if err == nil {
			a.loopSumInBytes = a.loopSumInBytes + requestBytes
		}
	}

	a.loopBeginTimestamp = time.Now().Unix()

	a.next.ServeHTTP(rw, req)

	responseBytesStr := rw.Header().Get("content-length")

	if responseBytesStr != "" {
		responseBytes, err := strconv.ParseInt(responseBytesStr, 10, 64)

		if err == nil {
			a.loopSumOutBytes = a.loopSumOutBytes + responseBytes
		}
	}
}
