package helpers

import (
	"net/http"
	"time"
)

// CreateHTTPClient creates an HTTP client with timeout configuration
func CreateHTTPClient(timeoutSeconds int) *http.Client {
	return &http.Client{
		Timeout: time.Duration(timeoutSeconds) * time.Second,
	}
}
