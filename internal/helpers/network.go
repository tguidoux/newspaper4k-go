package helpers

import (
	"net/http"
	"time"
)

const (
	// DefaultTimeoutSeconds is the default timeout for HTTP requests in seconds
	DefaultTimeoutSeconds int = 10
)

// CreateHTTPClient creates an HTTP client with timeout configuration
func CreateHTTPClient(timeoutSeconds int) *http.Client {
	return &http.Client{
		Timeout: time.Duration(timeoutSeconds) * time.Second,
	}
}

// CreateDefaultHTTPClient creates an HTTP client with default timeout
func CreateDefaultHTTPClient() *http.Client {
	return CreateHTTPClient(DefaultTimeoutSeconds)
}
