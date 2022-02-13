
package http_shaping

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"text/template"
	"time"
	"strconv"
	"errors"
	"strings"
	"unicode"
)

// Config the plugin configuration.
type Config struct {
	LoopInterval int64
	InTrafficLimit string
	OutTrafficLimit string
	ConsiderLimits bool
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		LoopInterval: 0,
	}
}

type HttpShaping struct {
	next     http.Handler
	name     string
	template *template.Template
	loopBeginTimestamp int64
	loopSumInBytes int64
	loopSumOutBytes int64
	loopInterval int64
	inTrafficLimit uint64
	outTrafficLimit uint64
	considerLimits bool
}

// New created a new HttpShaping plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	fmt.Print("Initiating http_shaping plugin ", name, "\n")
	fmt.Print("Config: ", config, "\n")

	if config.LoopInterval <= 0 {
		return nil, fmt.Errorf("LoopInterval should be positive")
	}

	inTrafficLimit, err := ToBytes(config.InTrafficLimit)

	if err != nil {
		return nil, fmt.Errorf("invalid inTrafficLimit")
	}

	outTrafficLimit, err := ToBytes(config.OutTrafficLimit)

	if err != nil {
		return nil, fmt.Errorf("invalid outTrafficLimit")
	}

	return &HttpShaping{
		next:     next,
		name:     name,
		template: template.New("httpShaping").Delims("[[", "]]"),
		loopBeginTimestamp: time.Now().Unix(),
		loopSumInBytes: 0,
		loopSumOutBytes: 0,
		loopInterval: config.LoopInterval,
		inTrafficLimit: inTrafficLimit,
		outTrafficLimit: outTrafficLimit,
		considerLimits: config.ConsiderLimits,
	}, nil
}

func (a *HttpShaping) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	fmt.Print("Serving http_shaping plugin, host=", req.Host, "\n")

	// reset the begin loop if needed
	if time.Now().Unix() - a.loopBeginTimestamp >= a.loopInterval {
		fmt.Print("Reseting the loop interval, host=", req.Host, "\n")
		a.loopBeginTimestamp = time.Now().Unix()
		a.loopSumOutBytes = 0
		a.loopSumInBytes = 0
	}

	requestBytesStr := req.Header.Get("content-length")

	if requestBytesStr != "" {
		requestBytes, err := strconv.ParseInt(requestBytesStr, 10, 64)

		if err == nil {
			a.loopSumInBytes = a.loopSumInBytes + requestBytes
		}
	}

	// check if we reached any of the limit
	fmt.Print("Current In: ", a.loopSumInBytes, " < ", a.inTrafficLimit, " host=", req.Host, "\n")

	if a.considerLimits && a.loopSumInBytes >= int64(a.inTrafficLimit) {
		http.Error(rw, "Bandwidth limit reached", http.StatusTooManyRequests)
		return
	}

	fmt.Print("Current Out: ", a.loopSumOutBytes, " < ", a.outTrafficLimit, " host=", req.Host, "\n")

	if a.considerLimits && a.loopSumOutBytes >= int64(a.outTrafficLimit) {
		http.Error(rw, "Bandwidth limit reached", http.StatusTooManyRequests)
		return
	}
	

	// forwarding the request
	a.next.ServeHTTP(rw, req)

	responseBytesStr := rw.Header().Get("content-length")

	if responseBytesStr != "" {
		responseBytes, err := strconv.ParseInt(responseBytesStr, 10, 64)

		if err == nil {
			a.loopSumOutBytes = a.loopSumOutBytes + responseBytes
		}
	}
}


const (
	BYTE = 1 << (10 * iota)
	KILOBYTE
	MEGABYTE
	GIGABYTE
	TERABYTE
	PETABYTE
	EXABYTE
)

var invalidByteQuantityError = errors.New("byte quantity must be a positive integer with a unit of measurement like M, MB, MiB, G, GiB, or GB")

// ToBytes parses a string formatted by ByteSize as bytes. Note binary-prefixed and SI prefixed units both mean a base-2 units
// KB = K = KiB	= 1024
// MB = M = MiB = 1024 * K
// GB = G = GiB = 1024 * M
// TB = T = TiB = 1024 * G
// PB = P = PiB = 1024 * T
// EB = E = EiB = 1024 * P
func ToBytes(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	s = strings.ToUpper(s)

	i := strings.IndexFunc(s, unicode.IsLetter)

	if i == -1 {
		return 0, invalidByteQuantityError
	}

	bytesString, multiple := s[:i], s[i:]
	bytes, err := strconv.ParseFloat(bytesString, 64)
	if err != nil || bytes < 0 {
		return 0, invalidByteQuantityError
	}

	switch multiple {
	case "E", "EB", "EIB":
		return uint64(bytes * EXABYTE), nil
	case "P", "PB", "PIB":
		return uint64(bytes * PETABYTE), nil
	case "T", "TB", "TIB":
		return uint64(bytes * TERABYTE), nil
	case "G", "GB", "GIB":
		return uint64(bytes * GIGABYTE), nil
	case "M", "MB", "MIB":
		return uint64(bytes * MEGABYTE), nil
	case "K", "KB", "KIB":
		return uint64(bytes * KILOBYTE), nil
	case "B":
		return uint64(bytes), nil
	default:
		return 0, invalidByteQuantityError
	}
}