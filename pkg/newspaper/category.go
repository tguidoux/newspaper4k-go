package newspaper

import (
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tguidoux/newspaper4k-go/internal/urls"
)

// Category represents a category object from a news source
type Category struct {
	URL  string
	HTML string
	Doc  *goquery.Document
}

// IsValidCategoryURL performs basic validation for category URLs
func IsValidCategoryURL(urlStr string) bool {
	if urlStr == "" {
		return false
	}

	parsedURL, err := urls.Parse(urlStr)
	if err != nil {
		return false
	}

	// Remove any URL that starts with #
	if strings.HasPrefix(parsedURL.Path, "#") {
		return false
	}

	// Remove URLs that are not http or https
	if parsedURL.Scheme != "" && parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false
	}

	pathChunks := urls.GetPathChunks(parsedURL.Path)

	// Remove index.html
	for i, chunk := range pathChunks {
		if chunk == "index.html" {
			pathChunks = append(pathChunks[:i], pathChunks[i+1:]...)
			break
		}
	}

	// We want a path with just one subdir
	if len(pathChunks) > 2 || len(pathChunks) == 0 {
		return false
	}

	if hasInvalidPrefixes(pathChunks) {
		return false
	}

	if len(pathChunks) == 2 && slices.Contains(CATEGORY_URL_PREFIXES, pathChunks[0]) {
		return true
	}

	return len(pathChunks) == 1 && len(pathChunks[0]) > 1 && len(pathChunks[0]) < 20
}

// hasInvalidPrefixes checks for invalid path prefixes
func hasInvalidPrefixes(pathChunks []string) bool {
	for _, chunk := range pathChunks {
		if strings.HasPrefix(chunk, "_") || strings.HasPrefix(chunk, "#") {
			return true
		}
	}
	return false
}
