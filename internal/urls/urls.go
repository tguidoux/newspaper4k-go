package urls

import (
	"net/url"
	"regexp"
	"slices"
	"strings"
)

const (
	// Date regex patterns for detecting dates in URLs
	dateRegexPattern      = `([\./\-_\s]?(19|20)\d{2})[\./\-_\s]?(([0-3]?[0-9][\./\-_\s])|(\w{3,5}[\./\-_\s]))([0-3]?[0-9]([\./\-\+\?]|$))`
	strictDateRegexPrefix = `(?<=\W)`
)

var (
	// Allowed file types for articles
	allowedTypes = []string{
		"html", "htm", "md", "rst", "aspx", "jsp", "rhtml", "cgi",
		"xhtml", "jhtml", "asp", "shtml",
	}

	// Good path keywords that indicate article content
	goodPaths = []string{
		"story", "article", "feature", "featured", "slides",
		"slideshow", "gallery", "news", "video", "media", "v",
		"radio", "press",
	}

	// Bad path chunks that indicate non-article content
	badChunks = []string{
		"careers", "contact", "about", "faq", "terms", "privacy",
		"advert", "preferences", "feedback", "info", "browse",
		"howto", "account", "subscribe", "donate", "shop", "admin",
		"auth_user", "emploi", "annonces", "blog", "courrierdeslecteurs",
		"page_newsletters", "adserver", "clicannonces", "services",
		"contribution", "boutique", "espaceclient",
	}

	// Bad domains to filter out
	badDomains = []string{
		"amazon", "doubleclick", "twitter", "facebook", "google",
		"youtube", "instagram", "pinterest",
	}

	// Compiled regex patterns
	dateRegex = regexp.MustCompile(dateRegexPattern)
)

// RedirectBack handles URL redirects from services like Pinterest
// that redirect to their site with the real news URL as a GET parameter
func RedirectBack(urlStr, sourceDomain string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}

	query := parsedURL.RawQuery

	// Parse query parameters
	queryValues, err := url.ParseQuery(query)
	if err != nil {
		return urlStr
	}

	// Check for 'url' parameter
	if urls := queryValues["url"]; len(urls) > 0 {
		return urls[0]
	}

	return urlStr
}

// PrepareURL cleans a URL, removes arguments, handles redirects,
// and merges relative URLs with absolute ones
func PrepareURL(urlStr string, sourceURL string) string {
	var properURL string

	if sourceURL != "" {
		sourceParsed, err := url.Parse(sourceURL)
		if err != nil {
			return ""
		}
		sourceDomain := sourceParsed.Host

		properURL = joinURL(sourceURL, urlStr)
		properURL = RedirectBack(properURL, sourceDomain)
	} else {
		properURL = urlStr
	}

	return properURL
}

// joinURL joins a base URL with a relative URL
func joinURL(baseURL, relativeURL string) string {
	base, err := url.Parse(baseURL)
	if err != nil {
		return relativeURL
	}

	rel, err := url.Parse(relativeURL)
	if err != nil {
		return relativeURL
	}

	return base.ResolveReference(rel).String()
}

// ValidURL checks if a URL is a valid news article URL
func ValidURL(urlStr string) bool {
	_, err := url.Parse(urlStr)
	return err == nil
}

func ValidArticleURL(urlStr string) bool {
	url, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	// Check for bad domains
	for _, badDomain := range badDomains {
		if strings.Contains(url.Host, badDomain) {
			return false
		}
	}

	// Check for bad path chunks
	pathChunks := strings.Split(url.Path, "/")
	for _, chunk := range pathChunks {
		for _, badChunk := range badChunks {
			if strings.EqualFold(chunk, badChunk) {
				return false
			}
		}
	}

	// Check for good path keywords
	goodPathFound := false
	for _, chunk := range pathChunks {
		for _, goodPath := range goodPaths {
			if strings.EqualFold(chunk, goodPath) {
				goodPathFound = true
				break
			}
		}
		if goodPathFound {
			break
		}
	}

	// Check for date patterns in the URL
	dateMatch := dateRegex.FindString(url.Path)
	if dateMatch != "" {
		goodPathFound = true
	}

	return goodPathFound
}

// URLToFileType extracts the file type from a URL
func URLToFileType(absURL string) string {
	parsedURL, err := url.Parse(absURL)
	if err != nil {
		return ""
	}

	path := strings.TrimSuffix(parsedURL.Path, "/")

	pathChunks := strings.Split(path, "/")
	if len(pathChunks) == 0 {
		return ""
	}

	lastChunk := pathChunks[len(pathChunks)-1]
	parts := strings.Split(lastChunk, ".")

	if len(parts) < 2 {
		return ""
	}

	fileType := parts[len(parts)-1]

	// Assume file extension is maximum 5 characters long
	if len(fileType) <= 5 || slices.Contains(allowedTypes, strings.ToLower(fileType)) {
		return strings.ToLower(fileType)
	}

	return ""
}

// GetDomain returns the domain part of a URL
func GetDomain(absURL string) string {
	parsedURL, err := url.Parse(absURL)
	if err != nil {
		return ""
	}
	return parsedURL.Host
}

// GetScheme returns the scheme part of a URL (http, https, ftp, etc)
func GetScheme(absURL string) string {
	parsedURL, err := url.Parse(absURL)
	if err != nil {
		return ""
	}
	return parsedURL.Scheme
}

// GetPath returns the path part of a URL
func GetPath(absURL string) string {
	parsedURL, err := url.Parse(absURL)
	if err != nil {
		return ""
	}
	return parsedURL.Path
}

// IsAbsURL checks if a URL is absolute
func IsAbsURL(urlStr string) bool {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	return parsedURL.Scheme != "" && parsedURL.Host != ""
}

// URLJoinIfValid joins a base URL and a possibly relative URL safely
func URLJoinIfValid(baseURL, relativeURL string) string {
	result := joinURL(baseURL, relativeURL)
	return result
}

type TLDData struct {
	Domain    string
	Subdomain string
}

// ExtractTLD extracts TLD information from a URL (simplified version)
func ExtractTLD(urlStr string) TLDData {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return TLDData{}
	}

	host := parsedURL.Host
	parts := strings.Split(host, ".")

	result := TLDData{}

	if len(parts) >= 2 {
		result.Domain = parts[len(parts)-2]
		if len(parts) >= 3 {
			result.Subdomain = strings.Join(parts[:len(parts)-2], ".")
		}
	}

	return result
}
